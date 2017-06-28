import socket
import sys


MAX_TRANSPORT_DATASIZE = 1024         

def readTransportData(host, port):

	tcp = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	orig = (host, port)
	tcp.bind(orig)
	tcp.listen(1)
	con, cliente = tcp.accept()
	msg = con.recv(MAX_TRANSPORT_DATASIZE).decode("utf-8")
	con.close()

	return msg

def sendNetworkData(host, port, msg):
	tcp = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	dest = (host, port)
	tcp.connect(dest)
	tcp.send(bytearray(msg,"utf-8"))

def str2intList(client_ip):

	client_ip = client_ip.split(".")
	i = 0
	for octect in client_ip:
		client_ip[i] = int(octect)
		i += 1	
	return client_ip

def intList2str(client_ip):
	ip = ""
	for octect in client_ip:
		ip += str(octect)
		ip += "."
	ip = ip[:len(ip)-1]

	return ip

def strList2str(transport_data):
	data = ""
	for item in transport_data:
		data += item
		data += "\n"
	return data

def sameNetwork(network_ip,client_ip,mask):

	ip = []
	i = 0
	for octect in client_ip:
		ip.append(octect & mask[i])
		i += 1
	i = 0
	for octect in network_ip:
		if ip[i] != octect:
			return False
		i += 1
	return True


HOST = '127.0.0.1'              
transport2network_port = int(sys.argv[1])
network2physical_port = 8011
physical2network_port = 8015
network2transport_port = 8016

#
# --------------------- IDA ----------------------------------
#

#
# Recebendo dados da camada de transporte
#
print("Aguardando dados da camada de transporte")
transport_data = readTransportData(HOST,transport2network_port)
print(transport_data)

#
# Variáveis para verificar se o IP dado pelo usuário está na suposta mesma rede
# do servidor
#
network_ip = [192,168,10,0]
mask = [255,255,255,0]
destination_ip = "192.168.10.15"

# 
# Enquanto o usuário não digitar um IP válido o programa não prossegue 
#
not_same_network = True
while(not_same_network):

	#
	# Menu para o usuário digitar IP do cliente
	#
	client_ip = input("Digite o IP da máquina cliente, mascara: 255.255.255.0\n> ")
	client_ip = str2intList(client_ip)

	if(sameNetwork(network_ip,client_ip,mask)):
		break
	print("Não é da mesma rede!")

#
# Montagem do cabeçalho da camada de Redes:
# IP de origem
# IP de destino
#
network_header = intList2str(client_ip) + "\n" + destination_ip + "\n"
network_data = network_header + transport_data


#
# Enviando mensagem da camada de rede pra camada física
#
print("Enviando mensagem da camada de rede pra camada física...")
sendNetworkData(HOST,network2physical_port,network_data)

#
# ---------------------- VOLTA ---------------------------------
#

#
# Usando a mesma função de leitura da camada fisica para ler a de transporte
#
readPhysicalData = readTransportData

#
# Receber dados da camada física
#
print("Aguardando dados da camada física...")
physical_data = readPhysicalData(HOST,physical2network_port)

#
# Retirando dados da camada de transporte
#
transport_data = physical_data.split("\n")[2:]
transport_data = strList2str(transport_data)

#
# Usando a mesma função de envio para a camada física
# para enviar dados para a camada de transporte
#
sendTransportData = sendNetworkData

#
# Enviando dados para camada de transporte
#
print("Enviando dados para camada de transporte...")
sendTransportData(HOST,network2transport_port,transport_data)