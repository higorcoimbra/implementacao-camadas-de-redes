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
//criando o socket
$socket = socket_create(AF_INET, SOCK_STREAM, 0) or die("Nao foi possivel criar o socket\n");
//ligando socket a porta
$valid = socket_bind($socket, $host, $physical2app_port) or die("Nao foi possivel ligar o socket a porta\n");
//começa a escutar por conexões na porta 8002
//o segundo parametro do socket_listen e o numero conexoes simultaneas nessa porta
$valid = socket_listen($socket, 1) or die("Nao foi possivel estabelecer a escuta do socket\n");

//aceita conexões na porta 8002
$spawn = socket_accept($socket) or die("Nao foi possivel conectar\n");
//le a mensagem de requisição HTTP da camada fisica do servidor
$mensagemHTTP = socket_read($spawn, $physical2app_port) or die("Nao foi possivel ler a entrada\n");
socket_close($socket);
socket_close($spawn);

/*
 * Transmitindo o arquivo solicitado a camada fisica
 */

list ($method, $filename) = split(' ',$mensagemHTTP);

echo $filename."\n";

$filename = str_replace("/","../",$filename);

$fp = fopen($filename, "r");

$content = fread($fp, filesize($filename));
$content = $content."\0";

$socket = socket_create(AF_INET, SOCK_STREAM, 0) or die("Nao foi possivel criar o socket\n");
$valid = socket_connect($socket, $host, $app2physical_port) or die ("Nao foi possivel conectar ao navegador");
$valid = socket_write($socket, $content) or die ("Nao foi possivel enviar mensagem");
socket_close($socket);

?>