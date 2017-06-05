<?php
$host = "127.0.0.1";
$http_port = $argv[1]; // Porta de comunicacao com o browser
$transport_app_port_communication = 8001;

/*
 * Recebe mensagem HTTP do navegador
 */

// Sem timeout!
set_time_limit(0);
// Criando o socket
$browser = socket_create(AF_INET, SOCK_STREAM, 0) or die("Nao foi possivel criar o socket\n");
// Ligando socket a porta
$valid = socket_bind($browser, $host, $http_port) or die("Nao foi possivel ligar o socket a porta\n");
echo "\n\nCamada de aplicacao do cliente aguardando requisicao do browser ...\n\n";
// Comeca a escutar por conexões na porta 8080
// o segundo parametro do socket_listen e' o numero conexoes simultaneas nessa porta
$valid = socket_listen($browser, 1) or die("Nao foi possivel estabelecer a escuta do socket\n");
// Aceita conexões na porta 8080
$spawn_browser = socket_accept($browser) or die("Nao foi possivel conectar\n");
// Le a mensagem de requisição HTTP do navegador
$mensagemHTTP = socket_read($spawn_browser, $http_port) or die("Nao foi possivel ler a entrada\n");
echo "Mensagem HTTP de requisicao recebida pelo browser:\n";
echo $mensagemHTTP."\n";

/*
 * Transmitindo mensagem HTTP a camada fisica
 */

echo "Envio da mensagem HTTP da camada de aplicacao cliente para camada de transporte do cliente\n\n";
$socket = socket_create(AF_INET, SOCK_DGRAM, 0) or die("Nao foi possivel criar o socket\n");
$valid = socket_connect($socket, $host, $transport_app_port_communication) or die ("Nao foi possivel conectar a camada de transporte\n");
$valid = socket_write($socket, $mensagemHTTP) or die ("Nao foi possivel enviar mensagem");
socket_close($socket);

/*
 * Recebendo mensagem de resposta HTTP da camada de transporte
 */

$socket = socket_create(AF_INET, SOCK_STREAM, 0) or die("Nao foi possivel criar o socket\n");
// Ligando socket a porta
$valid = socket_bind($socket, $host, $transport_app_port_communication) or die("Nao foi possivel ligar o socket a porta\n");
// Começa a escutar por conexões na porta 8005
//o segundo parametro do socket_listen e o numero conexoes simultaneas nessa porta
$valid = socket_listen($socket, 1) or die("Nao foi possivel estabelecer a escuta do socket\n");
// Aceita conexões na porta 8005
$spawn = socket_accept($socket) or die("Nao foi possivel conectar\n");
// Le a mensagem de resposta HTTP do buffer de entrada do cliente
$file = socket_read($spawn, $transport_app_port_communication) or die("Nao foi possivel ler a entrada\n");
socket_close($socket);
socket_close($spawn);
echo "Arquivo HTML recebido com sucesso do buffer de entrada\n\n";

echo $file;

/*
 * Transmitindo mensagem de resposta HTTP para o navegador
 */

echo "Transmitindo mensagem de resposta HTTP para o navegador\n\n";
socket_write($spawn_browser, $file);

socket_close($spawn_browser);

?>