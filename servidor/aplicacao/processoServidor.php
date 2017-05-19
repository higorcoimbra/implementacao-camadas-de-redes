<?php
$host = "127.0.0.1";
$http_port = 80;
$physical2app_port = 8002;
$app2physical_port = 8003;

/*
 * Recebe mensagem HTTP da camada fisica
 */

//sem timeout!
set_time_limit(0);
// Criando o socket
$socket = socket_create(AF_INET, SOCK_STREAM, 0) or die("Nao foi possivel criar o socket\n");
// Ligando socket a porta
$valid = socket_bind($socket, $host, $physical2app_port) or die("Nao foi possivel ligar o socket a porta\n");
// Comeca a escutar por conexões na porta 8002
//o segundo parametro do socket_listen e o numero conexoes simultaneas nessa porta
$valid = socket_listen($socket, 1) or die("Nao foi possivel estabelecer a escuta do socket\n");
// Aceita conexões na porta 8002
$spawn = socket_accept($socket) or die("Nao foi possivel conectar\n");
// Le a mensagem de requisição HTTP da camada fisica do servidor
$mensagemHTTP = socket_read($spawn, $physical2app_port) or die("Nao foi possivel ler a entrada\n");
socket_close($socket);
socket_close($spawn);
echo "\n\nMensagem HTTP recebida com sucesso do buffer de entrada do servidor\n\n";

/*
 * Transmitindo o arquivo solicitado a camada fisica
 */

$http_header="HTTP/1.1 200 OK
Connection: Keep-Alive
Content-Type: text/html

";
// Pega o nome do arquivo
list ($method, $filename) = preg_split('/ /',$mensagemHTTP);
// Abre o arquivo solicitado
$filename = str_replace("/","../",$filename);
$fp = fopen($filename, "r");
// Leitura do arquivo
$content = fread($fp, filesize($filename));
// Acrescenta o cabecalho http no arquivo solicitado
$content = $http_header.$content."\0";
echo "Conteudo da resposta HTTP:\n";
echo $content."\n\n";

// Envia arquivo html para o buffer de saida do servidor
echo "Envio do arquivo HTML para o buffer de saida do servidor\n\n";
$socket = socket_create(AF_INET, SOCK_STREAM, 0) or die("Nao foi possivel criar o socket\n");
$valid = socket_connect($socket, $host, $app2physical_port) or die ("Nao foi possivel conectar ao navegador");
$valid = socket_write($socket, $content) or die ("Nao foi possivel enviar mensagem");
socket_close($socket);

?>