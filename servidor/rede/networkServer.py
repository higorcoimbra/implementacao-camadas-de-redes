import socket
import sys

MAX_TRANSPORT_DATASIZE = 1024  

def readPhysicalData(host, port):

	
	

	return msg


def strList2str(transport_data):
	data = ""
	for item in transport_data:
		data += item
		data += "\n"
	return data

HOST = "127.0.0.1"
physical2network_port = 8005
network2transport_port = 8006
transport2network_port = 8009
network2physical_port = 8010

#
# --------------------- IDA -------------------------------
#
tcp = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
orig = (HOST, physical2network_port)
tcp.bind(orig)
tcp.listen(1)
while(1):
	#
	# Receber dados da camada física
	#
	print("Aguardando dados da camada física...")
	con, cliente = tcp.accept()
	msg = con.recv(MAX_TRANSPORT_DATASIZE).decode("utf-8")
	
	physical_data = msg

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
	tcp2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	dest = (HOST, network2transport_port)
	tcp2.connect(dest)
	tcp2.send(bytearray(transport_data,"utf-8"))
	tcp2.close()

	if(physical_data.find("LASTSEG") > -1):
		break
con.close()
#
# --------------------- VOLTA ----------------------------
#

print("\n---Aguardando a volta---\n")

tcp = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

orig = (HOST, transport2network_port)
tcp.bind(orig)
tcp.listen(1)

while(1):
	#
	# Recebendo dados da camada de transporte
	#
	print("Aguardando dados da camada de transporte")
	con, cliente = tcp.accept()
	transport_data = con.recv(MAX_TRANSPORT_DATASIZE).decode("utf-8")
	print(transport_data)


	#
	# Montagem do cabeçalho da camada de Redes:
	# IP de origem
	# IP de destino
	#
	network_header = ips[1] + "\n" + ips[0] + "\n"
	network_data = network_header + transport_data

	#
	# Enviando mensagem da camada de rede pra camada física
	#
	print("Enviando mensagem da camada de rede pra camada física...")
	tcp2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	dest = (HOST, network2physical_port)
	tcp2.connect(dest)
	tcp2.send(bytearray(network_data,"utf-8"))
	tcp2.close()
	if(transport_data.find("LASTSEG") > -1):
		break

con.close()