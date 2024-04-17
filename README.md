# Online chat

Projeto de chat em grupo, utilizando grpc.


## Server

Server implementa os dois métodos: 

```proto
rpc UpdateUserStatus(UserStatusChange) returns (stream ChatMessage) {}
rpc SendMessage(ChatMessage) returns (ChatMessageResponse) {}
```
Server implementa um array de canais `pb.ChatMessage`.

- UpdateUserStatus: Cria um novo canal, representando a conexão `grpc`, na lista compartilhada de canais. Em seguida, cria uma go routine, que implementa um `loop` de leitura no canal recém criado. Quando o canal receber uma mensagem, o método encaminha a mensagem para o `client`, via `stream`. 

- SendMessage: Client cria uma `request`, a mensagem vai ser escrita em todos os canais na estrutura `server`. Retorna `ok=true` em caso de sucesso.

> Observação: Essa primeira versão não leva em consideração nenhum [pattern de concorrência](https://go.dev/blog/pipelines) já bem estruturado no go.  


## Proto

O arquivo `.proto` está na raiz, para compilar rodar o comando:

```proto
protoc --go_out=proto --go_opt=paths=source_relative --go-grpc_out=proto --go-grpc_opt=paths=source_relative .\online-chat.proto
```

## Desenho

[image](./images/online-chat-grpc.png)