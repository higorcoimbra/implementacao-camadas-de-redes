import socket
import sys

MAX_TRANSPORT_DATASIZE = 1024  

def readPhysicalData(host, port):

	tcp = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	orig = (host, port)
	tcp.bind(orig)
	tcp.listen(1)
	con, cliente = tcp.accept()
	msg = con.recv(MAX_TRANSPORT_DATASIZE).decode("utf-8")
	con.close()

	return msg

def sendTransportData(host, port, msg):
	tcp = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	dest = (host, port)
	tcp.connect(dest)
	tcp.send(bytearray(msg,"utf-8"))

def strList2str(transport_data):
	data = ""
	for item in transport_data:
		data += item
		data += "\n"
	return data

HOST = "127.0.0.1"
physical2network_port = int(sys.argv[1])
network2transport_port = 8012
transport2network_port = 8013
network2physical_port = 8014

#
# --------------------- IDA -------------------------------
#

#
# Receber dados da camada física
#
print("Aguardando dados da camada física...")
physical_data = readPhysicalData(HOST,physical2network_port)

#
# Retirando os ips da camada de rede
#
ips = physical_data.split("\n")[0:2]

#
# Retirando dados da camada de transporte
#
transport_data = physical_data.split("\n")[2:]
transport_data = strList2str(transport_data)

#
# Enviando dados para camada de transporte
#
print("Enviando dados para a camada de transporte...")
sendTransportData(HOST,network2transport_port,transport_data)

#
# --------------------- VOLTA ----------------------------
#

#
# Usando a mesma função de leitura da camada fisica para ler a de transporte
#
readTransportData = readPhysicalData

#
# Lendo a resposta HTTP da camada de transporte
#
print("Aguardando resposta HTTP da camada de transporte...")
transport_data = readTransportData(HOST,transport2network_port)

#
# Montagem do cabeçalho da camada de Redes:
# IP de origem
# IP de destino
#
network_header = ips[1] + "\n" + ips[0] + "\n"
network_data = network_header + transport_data
print(network_data)

#
# Usando a mesma função de envio para a camada de transporte
# para enviar dados para a camada física
#
sendNetworkData = sendTransportData

#
# Enviando mensagem da camada de rede pra camada física
#
print("Enviando dados da camada de rede para camada física...")
sendNetworkData(HOST,network2physical_port,network_data)



