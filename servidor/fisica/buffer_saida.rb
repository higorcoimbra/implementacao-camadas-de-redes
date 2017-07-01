=begin
	Camada física
	Cliente
=end

require 'socket'
require 'macaddr'


HEXA = Hash['1'=>'0001','2'=>'0010','3'=>'0011','4'=>'1000',
			'5'=>'0101','6'=>'0110','7'=>'0111','8'=>'1000',
			'9'=>'1001','a'=>'1010','b'=>'1011','c'=>'1100',
			'd'=>'1101','e'=>'1110','f'=>'1111','0'=>'0000']


# Transforma o mac address em macAddr_binario
def getMacAddrBit(mac)
	macAddr_binario = ""
	mac.each_char do |i|
		if (i != ":" and HEXA[i] != nil)
			macAddr_binario += HEXA[i]
		end
	end
	return macAddr_binario
end


# Conexao TCP de envio de mensagem
def tcpSendConnect(host, port, mensagem)
	server = TCPSocket.open(host, port)
	server.puts(mensagem)
	server.close()
end

# Recebe pacote da camada de rede
def rcv_pkt(network2physical_port, conexao)
	application = conexao.accept
	mensagem = application.read()
	return mensagem
end


def makeCamadaFisicaHeaderTCP(quadro, preambulo, sof, macAddrClientBit, macAddrServerBit, etherType)
	quadro.print(preambulo)
	quadro.print(sof)
	quadro.print(macAddrClientBit)
	quadro.print(macAddrServerBit)
	quadro.print(etherType)
	return quadro
end


def probabilidadeColisao(quadro_index)
	colisao = rand(1...100) > 70
	while colisao
		puts("        --- PROBABILIDADE DE COLISAO ---")
		puts("Ocorreu colisao no envio do quadro "+quadro_index.to_s+"\n")
		sleep(rand(1...100)/100)
		puts("Reenvio do quadro "+quadro_index.to_s+"\n\n")
		colisao = rand(1...100) > 70
	end
end


# Variaveis de configuracao de host
host = '127.0.0.1'
port = 8011
network2physical_port = 8010
macAddrClient = Mac.addr
macAddrServer = 'aa:aa:aa:aa:aa:aa'

# Variaveis de configuracao da transmissao
transmissionTMQClient = 100
tamanho_infinito = 0x3f3f3f3f
transmissionTMQServer = tamanho_infinito
transmissionTMQ = transmissionTMQClient


# Definicao do cabecalho

=begin

	Estrutura do quadro a ser enviado

	[ Preambulo - SOF - MAC Destino - MAC Origem - Ether type - Dados ]
	IPOrigem
	IPDestino
	destinationPort sourcePort
	sequencenumber
	ack
	data
	TRAILER/LASTSEG

=end

headerSize = 22
preambulo="10101010101010101010101010101010101010101010101010101010"
sof="10101011"
macAddrClientBit = getMacAddrBit(macAddrClient)
macAddrServerBit = getMacAddrBit(macAddrServer)
etherType = "0000"


#
# ----------- CONEXAO COM CAMADA FISICA DO SERVIDOR - DEFINIR TMQ ----------------------
#


# Definicao do TMQ
server = TCPSocket.open(host, port)
transmissionTMQServer = Integer(server.gets)
server.close()
puts("         --- DIFINICAO TMQ ---")
puts("\nTMQ do Cliente: "+transmissionTMQClient.to_s+" bytes\n")
puts("TMQ do Servidor: "+transmissionTMQServer.to_s+" bytes\n")
transmissionTMQ = transmissionTMQClient < transmissionTMQServer ? transmissionTMQClient : transmissionTMQServer
puts("TMQ definido: "+transmissionTMQ.to_s+" bytes\n\n")

# Verifica se existe espaco para os dados
dataSize = transmissionTMQ - headerSize
if dataSize < 1
	puts("Largura de Banda Insuficiente. ")
	exit 1
end

# Pega informacoes sobre MAC Address do servidor
server = TCPSocket.open(host, port)
macAddrServer = server.gets
server.close()


#
# ----------- ESCREVENDO QUADRO E ENVIANDO ----------------------
#


conexao = TCPServer.open(network2physical_port)
quadro_index = 1
vestigio = ""
isLASTSEG = false
transferencia_aberta = true
while transferencia_aberta

	# - Cria quadro

	# Escreve cabecalho
	puts("---------------------------------------")
	puts("Enviando quadro "+quadro_index.to_s+" ...\n\n")
	quadro = File.new("quadro.txt", "w")
	quadro = makeCamadaFisicaHeaderTCP(quadro, preambulo, sof, macAddrClientBit, macAddrServerBit, etherType)
	puts("               --- CABECALHO DO QUADRO ---")
	puts("Preambulo: "+preambulo+"\n")
	puts("SOF: "+sof+"\n")
	puts("MAC address do cliente: "+macAddrClient.to_s+"\n")
	puts("MAC address do servidor: "+macAddrServer.to_s)
	puts("Ether Type: "+etherType+"\n\n")

	# Escreve dados no quadro
	j = 0
	for p in 0..(vestigio.length-1)
		j += 1
		quadro.print(vestigio[p].ord.to_s(2).rjust(10, '0'))
		if vestigio[vestigio.length-7..vestigio.length-1] == "LASTSEG"
			isLASTSEG = true
			transferencia_aberta = false
		end
	end

	vestigio = ""
	if not isLASTSEG
		while j < dataSize
			pkt = rcv_pkt(network2physical_port, conexao)
			puts("         --- PACOTE RECEBIDO DA CAMADA DE REDE ---")
			print(pkt)
			puts("")
			for p in 0..(pkt.length-1)
				if j >= dataSize
					vestigio = pkt[p..pkt.length]
					break
				end
				quadro.print(pkt[p].ord.to_s(2).rjust(10, '0'))
				j += 1
			end

			if pkt[pkt.length-7..pkt.length-1] == "LASTSEG"
				if vestigio == ""
					transferencia_aberta = false
				end
				break
			end
		end
	end

	quadro.close()
	isLASTSEG = false

	# - Ler dados do quadro
	pdu = File.read('quadro.txt')

	# - Probabilidade de colisao
	probabilidadeColisao(quadro_index)

	# - Envio da pdu
	tcpSendConnect(host, port, pdu)
	puts("Envio do quador "+quadro_index.to_s+" realizado com sucesso\n")
	quadro_index += 1

end
