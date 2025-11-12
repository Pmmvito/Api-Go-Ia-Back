# Script para testar erro 404 quando email não existe

Write-Host "`n===== TESTE: Recuperação de Senha - Email Não Cadastrado =====`n" -ForegroundColor Cyan

# Email que não existe no banco
$emailNaoCadastrado = "email-nao-existe@example.com"

$body = @{
    email = $emailNaoCadastrado
} | ConvertTo-Json

Write-Host "Tentando recuperar senha para email não cadastrado:" -ForegroundColor Yellow
Write-Host "Email: $emailNaoCadastrado" -ForegroundColor Yellow
Write-Host ""

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/forgot-password" `
        -Method POST `
        -Body $body `
        -ContentType "application/json" `
        -ErrorAction Stop
    
    Write-Host "ERRO: Deveria ter retornado 404!" -ForegroundColor Red
    $response | ConvertTo-Json -Depth 10
    
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    
    if ($statusCode -eq 404) {
        Write-Host "✅ SUCESSO! Retornou erro 404 como esperado" -ForegroundColor Green
        Write-Host ""
        Write-Host "Status Code: $statusCode" -ForegroundColor Green
        
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $reader.BaseStream.Position = 0
            $reader.DiscardBufferedData()
            $responseBody = $reader.ReadToEnd()
            
            Write-Host "Resposta do servidor:" -ForegroundColor Green
            Write-Host $responseBody -ForegroundColor White
            
            # Parse JSON para verificar estrutura
            $jsonResponse = $responseBody | ConvertFrom-Json
            Write-Host ""
            Write-Host "Estrutura esperada:" -ForegroundColor Cyan
            Write-Host "  - message: $($jsonResponse.message)" -ForegroundColor White
        }
    } else {
        Write-Host "❌ ERRO: Status code incorreto!" -ForegroundColor Red
        Write-Host "Esperado: 404" -ForegroundColor Red
        Write-Host "Recebido: $statusCode" -ForegroundColor Red
        
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $reader.BaseStream.Position = 0
            $reader.DiscardBufferedData()
            $responseBody = $reader.ReadToEnd()
            Write-Host "Resposta: $responseBody" -ForegroundColor Yellow
        }
    }
}

Write-Host "`n================================================================`n" -ForegroundColor Cyan
