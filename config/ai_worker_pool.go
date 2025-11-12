package config

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// AIJob representa um job de categorização com IA
type AIJob struct {
	ID       string
	UserID   uint
	Items    interface{}
	Callback func(result interface{}, err error)
	Context  context.Context
}

// AIWorkerPool gerencia o processamento de requisições para a IA
// Otimizado para Gemini 2.5 Flash Preview (Free Tier):
// - Limite: ~10 RPM (requests por minuto)
// - Recomendado: 3 workers simultâneos para segurança
type AIWorkerPool struct {
	maxWorkers     int
	queue          chan AIJob
	semaphore      chan struct{}
	wg             sync.WaitGroup
	stats          AIStats
	mu             sync.RWMutex
	rateLimiter    *time.Ticker
	shutdownChan   chan struct{}
	isShuttingDown bool
}

// AIStats estatísticas do pool de workers
type AIStats struct {
	TotalProcessed    int64
	TotalFailed       int64
	TotalQueued       int64
	CurrentInQueue    int
	CurrentProcessing int
	TotalTime         time.Duration
}

var (
	aiPool   *AIWorkerPool
	poolOnce sync.Once
)

// InitAIWorkerPool inicializa o pool de workers para IA
// Para Gemini 2.5 Flash Preview Free:
// - maxWorkers: 3 (recomendado para 10 RPM)
// - queueSize: 50 (buffer de jobs esperando)
func InitAIWorkerPool(maxWorkers, queueSize int) {
	poolOnce.Do(func() {
		aiPool = &AIWorkerPool{
			maxWorkers:   maxWorkers,
			queue:        make(chan AIJob, queueSize),
			semaphore:    make(chan struct{}, maxWorkers),
			rateLimiter:  time.NewTicker(6 * time.Second), // 10 req/min = 1 req a cada 6s
			shutdownChan: make(chan struct{}),
		}
		aiPool.start()
		log.Printf("🤖 AI Worker Pool iniciado: %d workers, fila de %d, rate limit: 10 req/min", maxWorkers, queueSize)
	})
}

// GetAIWorkerPool retorna a instância singleton do pool
func GetAIWorkerPool() *AIWorkerPool {
	return aiPool
}

// start inicia os workers do pool
func (p *AIWorkerPool) start() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker processa jobs da fila com rate limiting
func (p *AIWorkerPool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.shutdownChan:
			log.Printf("🛑 Worker %d: Encerrando...", id)
			return
		case job, ok := <-p.queue:
			if !ok {
				log.Printf("🛑 Worker %d: Fila fechada, encerrando...", id)
				return
			}

			// Adquire slot do semáforo
			p.semaphore <- struct{}{}
			p.incrementProcessing()
			p.decrementQueue()

			// Rate limiting: aguarda ticker para respeitar 10 RPM
			<-p.rateLimiter.C

			start := time.Now()
			log.Printf("🤖 Worker %d: Processando job %s (user %d)", id, job.ID, job.UserID)

			// Verifica se contexto foi cancelado
			select {
			case <-job.Context.Done():
				log.Printf("⚠️ Worker %d: Job %s cancelado", id, job.ID)
				if job.Callback != nil {
					job.Callback(nil, fmt.Errorf("job cancelado"))
				}
				p.incrementFailed()
			default:
				// Executa callback (sua função de IA)
				if job.Callback != nil {
					job.Callback(job.Items, nil)
				}
			}

			duration := time.Since(start)
			p.incrementProcessed()
			p.addTime(duration)

			log.Printf("✅ Worker %d: Job %s concluído em %v", id, job.ID, duration)

			// Libera slot do semáforo
			p.decrementProcessing()
			<-p.semaphore
		}
	}
}

// SubmitJob adiciona um job à fila de processamento
func (p *AIWorkerPool) SubmitJob(job AIJob) error {
	if p.isShuttingDown {
		return fmt.Errorf("worker pool está encerrando, não aceita novos jobs")
	}

	select {
	case p.queue <- job:
		p.incrementQueue()
		log.Printf("📥 Job %s adicionado à fila (fila: %d)", job.ID, p.GetQueueSize())
		return nil
	default:
		return fmt.Errorf("fila de processamento cheia (%d jobs). Tente novamente em alguns minutos", cap(p.queue))
	}
}

// GetStats retorna estatísticas do pool
func (p *AIWorkerPool) GetStats() AIStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	stats := p.stats
	stats.CurrentInQueue = len(p.queue)
	return stats
}

// GetQueueSize retorna o tamanho atual da fila
func (p *AIWorkerPool) GetQueueSize() int {
	return len(p.queue)
}

// GetQueueCapacity retorna a capacidade máxima da fila
func (p *AIWorkerPool) GetQueueCapacity() int {
	return cap(p.queue)
}

// IsQueueFull verifica se a fila está cheia
func (p *AIWorkerPool) IsQueueFull() bool {
	return len(p.queue) >= cap(p.queue)
}

// Shutdown encerra o pool gracefully
func (p *AIWorkerPool) Shutdown(timeout time.Duration) error {
	p.isShuttingDown = true
	close(p.shutdownChan)

	// Aguarda workers finalizarem com timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("✅ AI Worker Pool encerrado com sucesso")
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout ao encerrar worker pool")
	}
}

// Métodos auxiliares para estatísticas
func (p *AIWorkerPool) incrementProcessed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stats.TotalProcessed++
}

func (p *AIWorkerPool) incrementFailed() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stats.TotalFailed++
}

func (p *AIWorkerPool) incrementProcessing() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stats.CurrentProcessing++
}

func (p *AIWorkerPool) decrementProcessing() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stats.CurrentProcessing--
}

func (p *AIWorkerPool) incrementQueue() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stats.TotalQueued++
}

func (p *AIWorkerPool) decrementQueue() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stats.CurrentInQueue--
	if p.stats.CurrentInQueue < 0 {
		p.stats.CurrentInQueue = 0
	}
}

func (p *AIWorkerPool) addTime(duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stats.TotalTime += duration
}

// GetAverageProcessingTime retorna o tempo médio de processamento
func (p *AIWorkerPool) GetAverageProcessingTime() time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.stats.TotalProcessed == 0 {
		return 0
	}
	return p.stats.TotalTime / time.Duration(p.stats.TotalProcessed)
}
