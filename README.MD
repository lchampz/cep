# API de Distância entre CEPs

Esta API calcula a distância entre dois CEPs brasileiros e verifica se estão dentro de um raio especificado.

## Requisitos

- Go 1.16 ou superior
- Dependências:
  - github.com/gin-gonic/gin

## Instalação

1. Clone o repositório
2. Instale as dependências:
```bash
go mod tidy
```
3. Execute o servidor:
```bash
go run main.go
```

O servidor estará rodando em `http://localhost:8080`

## Endpoints

### Calcular Distância entre CEPs

**Endpoint:** `POST /cep`

**Request Body:**
```json
{
    "from": "12345678",
    "to": "87654321",
    "range": 10.0
}
```

| Campo | Tipo   | Descrição                                    |
|-------|--------|----------------------------------------------|
| from  | string | CEP de origem (8 dígitos)                    |
| to    | string | CEP de destino (8 dígitos)                   |
| range | number | Distância máxima em quilômetros              |

**Response:**
```json
{
    "km": "5.23",
    "inRange": true
}
```

| Campo    | Tipo    | Descrição                                    |
|----------|---------|----------------------------------------------|
| km       | string  | Distância calculada em quilômetros           |
| inRange  | boolean | Indica se os CEPs estão dentro do raio       |

## Exemplos de Uso

### cURL
```bash
curl -X POST http://localhost:8080/cep \
-H "Content-Type: application/json" \
-d '{
    "from": "12345678",
    "to": "87654321",
    "range": 10.0
}'
```

### Python
```python
import requests

url = "http://localhost:8080/cep"
data = {
    "from": "12345678",
    "to": "87654321",
    "range": 10.0
}

response = requests.post(url, json=data)
print(response.json())
```

### JavaScript
```javascript
fetch('http://localhost:8080/cep', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify({
        from: '12345678',
        to: '87654321',
        range: 10.0
    })
})
.then(response => response.json())
.then(data => console.log(data));
```

## Códigos de Erro

| Código | Descrição                                    |
|--------|----------------------------------------------|
| 400    | Erro na validação dos dados de entrada       |
| 500    | Erro interno do servidor                     |

## Observações

- Os CEPs devem conter exatamente 8 dígitos numéricos
- A API utiliza o serviço Nominatim para geocodificação
- O cálculo de distância é feito usando a fórmula de Haversine
- Existe um limite de requisições para o serviço de geocodificação
- Os logs são salvos no arquivo `app.log`

## Limitações

- A API só funciona com CEPs brasileiros
- O serviço de geocodificação pode ter latência
- Existe um timeout de 5 segundos para requisições
- Máximo de 3 tentativas em caso de falha
