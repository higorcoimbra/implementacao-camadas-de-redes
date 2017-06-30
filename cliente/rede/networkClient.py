import socket
import sys


MAX_TRANSPORT_DATASIZE = 1024         


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

def belongsToNet(network_ip,client_ip,mask):

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

def sameNetwork(client_ip,destination_ip):
	return belongsToNet(network_ip,str2intList(client_ip),mask) and belongsToNet(network_ip,str2intList(destination_ip),mask)
	

HOST = '127.0.0.1'              
transport2network_port = 8002
network2physical_port = 8003
physical2network_port = 8012
network2transport_port = 8013


network_ip = [192,168,10,0]
mask = [255,255,255,0]
destination_ip = "192.168.10.15"
client_ip = "192.168.10.10"

print("****** CHECANDO SE OS IPS DE CLIENTE E SERVIDOR ESTÃO NA MESMA REDE ******")
if(sameNetwork(client_ip,destination_ip)):
	print("IPs pertencem a mesma rede!")
	print("\n")
else:
	print("IPs não pertencem a mesma rede, pacote será redirecionado a um roteador!")
	print("\n")
print("**************************************************************************")

#
# --------------------- IDA ----------------------------------
#

print("*************** INICIANDO FLUXO DE REQUISIÇÃO AO SERVIDOR ***************")
print("\n")

tcp = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
orig = (HOST, transport2network_port)
tcp.bind(orig)
tcp.listen(1)

while(1):
	#
	# Recebendo dados da camada de transporte
	#
	print("Aguardando segmento da requisição HTTP da camada de transporte...")
	print("\n")
	con, cliente = tcp.accept()
	transport_data = con.recv(MAX_TRANSPORT_DATASIZE).decode("utf-8")
	print("***** SEGMENTO RECEBIDO *****")
	print(transport_data)
	print("*****************************")
	print("\n")

	#
	# Montagem do cabeçalho da camada de Redes:
	# IP de origem
	# IP de destino
	#
	network_header = client_ip + "\n" + destination_ip + "\n"
	network_data = network_header + transport_data

	print("****** DATAGRAMA MONTADO ******")
	print(network_data)
	print("*******************************")
	print("\n")

	#
	# Enviando datagrama pra camada física
	#
	print("Enviando datagrama da requisição HTTP pra camada física...")
	print("\n")
	tcp2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	dest = (HOST, network2physical_port)
	tcp2.connect(dest)
	tcp2.send(bytearray(network_data,"utf-8"))
	tcp2.close()
	if(transport_data.find("LASTSEG") > -1):
		break

con.close()

#
# ---------------------- VOLTA ---------------------------------
#

print("*************** INICIANDO FLUXO DE RESPOSTA DO SERVIDOR ***************")
print("\n")

tcp = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
orig = (HOST, physical2network_port)
tcp.bind(orig)
tcp.listen(1)
while(1):
	#
	# Receber dados da camada física
	#
	print("Aguardando quadro da resposta HTTP da camada física...")
	print("\n")
	con, cliente = tcp.accept()
	msg = con.recv(MAX_TRANSPORT_DATASIZE).decode("utf-8")
	physical_data = msg

	print("****** QUADRO RECEBIDO ******")
	print(physical_data)
	print("*****************************")
	print("\n")

	#
	# Retirando cabeçalho da camada de rede dos dados da camada de transporte
	#
	transport_data = physical_data.split("\n")[2:]
	transport_data = strList2str(transport_data)

	print("****** SEGMENTO DESENCAPSULADO ******")
	print(transport_data)
	print("*************************************")
	print("\n")

	#
	# Enviando dados para camada de transporte
	#
	print("Enviando segmento da resposta HTTP para a camada de transporte...")
	print("\n")
	tcp2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	dest = (HOST, network2transport_port)
	tcp2.connect(dest)
	tcp2.send(bytearray(transport_data,"utf-8"))
	tcp2.close()

	if(physical_data.find("LASTSEG") > -1):
		break
con.close()