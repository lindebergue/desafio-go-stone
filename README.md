# desafio-go-stone
REST API de transferência entre contas de um banco digital, feita para o desafio
técnico Golang da Stone.

## Começando
O `docker-compose.yml` contém tudo que é necessário para rodar o projeto. Basta executar `docker-compose up` e aguardar a inicialização que a aplicação estará disponível na porta 9999.

No `docker-compose.yml` é definido os contêineres do projeto (app) e do banco de dados Postgres (db). A aplicação pode ser rodada diretamente fora do docker executando `go run .` na raiz do projeto e definindo as variáveis de ambiente a seguir:

| Variável | Descrição |
|----------|-----------|
| `APP_JWT_SECRET` | Chave secreta usada para gerar JSON Web Tokens |
| `APP_DATABASE_URL` | URL de conexão com o banco de dados no formato `postgres://user:pass@host:port/db` |

## Estrutura do projeto
O projeto implementa a API a partir do arquivo `main.go` com os pacotes abaixo:

- `auth/` - Funções e rotinas de criação de hashes e tokens
- `database/` - Camada de acesso de banco de dados
- `router/` - Rotas HTTP da aplicação

Cada um dos pacotes possui testes unitários padrão do Golang, executáveis com `go test`. Para executar os testes de integração com o banco de dados, defina a variável de ambiente `DATABASE_URL` com a URL correspondente. Se não for definida, os testes usam um mock do banco de dados com os dados armazenados na memória.
