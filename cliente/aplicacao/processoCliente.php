<?php
$host = "127.0.0.1";
$http_port = 80;
$app2physical_port = 8001;

/*
 * Recebe mensagem HTTP do navegador
 */

//sem timeout!
set_time_limit(0);
//criando o socket
$socket = socket_create(AF_INET, SOCK_STREAM, 0) or die("Nao foi possivel criar o socket\n");
//ligando socket a porta
$valid = socket_bind($socket, $host, $http_port) or die("Nao foi possivel ligar o socket a porta\n");
//começa a escutar por conexões na porta 80
//o segundo parametro do socket_listen e o numero conexoes simultaneas nessa porta
$valid = socket_listen($socket, 1) or die("Nao foi possivel estabelecer a escuta do socket\n");

//aceita conexões na porta 80
$spawn = socket_accept($socket) or die("Nao foi possivel conectar\n");
//le a mensagem de requisição HTTP do navegador
$mensagemHTTP = socket_read($spawn, $http_port) or die("Nao foi possivel ler a entrada\n");
socket_close($socket);
socket_close($spawn);

/*
 * Transmitindo mensagem HTTP a camada fisica
 */

$socket = socket_create(AF_INET, SOCK_STREAM, 0) or die("Nao foi possivel criar o socket\n");
$valid = socket_connect($socket, $host, $app2physical_port) or die ("Nao foi possivel conectar ao navegador");
$valid = socket_write($socket, $mensagemHTTP) or die ("Nao foi possivel enviar mensagem");
socket_close($socket);


?>