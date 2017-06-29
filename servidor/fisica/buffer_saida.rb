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


# Transforma o mac address em binario
def getMacBit(mac)
	binario = ""
	mac.each_char do |i|
		if (i != ":" and HEXA[i] != nil)
			binario += HEXA[i]
		end
	end
	return binario
end


# Conexao TCP e envio de mensagem
def tcpConnect(host, port, mensagem)
	server = TCPSocket.open(host, port)
	server.puts(mensagem)
	server.close()
end

def rcv_pkt(transport2physical_port, interface)
	#Cria um socket para transmissao da mensagem HTTP do processo cliente da aplicacao (navegador)..
	#..para a camada fisica
	print("Aguardando conexao da camada de transporte na porta ", transport2physical_port," ...")
	#interface = TCPServer.open(transport2physical_port)
	application = interface.accept
	print("Aguardando recebimento de pacote na porta ", transport2physical_port)
	mensagem = application.read()
	puts("\n\nSegmento recebido com sucesso da camada de transporte cliente\n\n")
	#interface.close()
	return mensagem
end


def makeCamadaFisicaHeaderTCP(pacote, preambulo, sof, macClientBit, macServerBit, etherType)
	pacote.print(preambulo)
	pacote.print(sof)
	pacote.print(macClientBit)
	pacote.print(macServerBit)
	pacote.print(etherType)
	return pacote
end


# Variaveis uteis
infinito = 0x3f3f3f3f
headerSize = 22

# Variaveis de configuracao de host
host = '127.0.0.1'
port = 8004
transport2physical_port = 8008
macClient = Mac.addr
macServer = 'aa:aa:aa:aa:aa:aa'

# Variaveis de configuracao da transmissao
transmissionClient = 120
transmissionServer = infinito
gargalo = transmissionClient

# Cria uma conexao TCP para pegar maximo de transmissao
server = TCPSocket.open(host, port)
transmissionServer = Integer(server.gets)
server.close()
puts("Estabelecimento de conexao com o buffer de entrada do servidor ...\n")
puts("\nTMP do Cliente: "+transmissionClient.to_s+" bytes\n")
puts("TMP do Servidor: "+transmissionServer.to_s+" bytes\n")
gargalo = transmissionClient < transmissionServer ? transmissionClient : transmissionServer
puts("TMP definido: "+gargalo.to_s+" bytes\n\n")

# Verifica se existe espaco para os dados
dataSize = gargalo - headerSize
if dataSize < 1
	puts("Largura de Banda Insuficiente. ")
	tcpConnect(host, port, "acabou")
	exit 1
end

# Pega informacoes de host do cliente
server = TCPSocket.open(host, port)
macServer = server.gets
server.close()

# Envio dos pacotes - Transformacao em binario, formatacao e envio
# [ Preambulo - SOF - MAC Destino - MAC Origem - Ether type - Dados ]

preambulo="10101010101010101010101010101010101010101010101010101010"
sof="10101011"
macClientBit = getMacBit(macClient)
macServerBit = getMacBit(macServer)
etherType = "0000"
transferencia_aberta = true

=begin
	destinationPort sourcePort
	sequencenumber
	ack
	data
	TRAILER

	No destino temos que ir lendo os bytes
	Assim que ler um TRAILER, significa que acabamos de ler um pacote
	entao enviamos esse pacote para a camada superior
	Desta forma mantemos a estrutura atual de envio da 
=end

package_index = 1
vestigio = ""
last_seg = false

interface = TCPServer.open(transport2physical_port)


while transferencia_aberta
	# Cria pacote
	# - Escreve cabecalho
	puts("---------------------------------------")
	puts("Enviando quadro "+package_index.to_s+" ...")
	pacote = File.new("pacote.txt", "w")
	pacote = makeCamadaFisicaHeaderTCP(pacote, preambulo, sof, macClientBit, macServerBit, etherType)
	puts("Preambulo: "+preambulo+"\n")
	puts("SOF: "+sof+"\n")
	puts("MAC address do cliente: "+macClient.to_s+"\n")
	puts("MAC address do servidor: "+macServer.to_s)
	puts("Ether Type: "+etherType+"\n\n")

	# - Escreve dados no pacote
	j = 0
	for p in 0..(vestigio.length-1)
		j += 1
		pacote.print(vestigio[p].ord.to_s(2).rjust(10, '0'))
		if vestigio[vestigio.length-7..vestigio.length-1] == "LASTSEG"
			last_seg = true
			transferencia_aberta = false
		end
	end

	vestigio = ""
	if not last_seg
		while j < dataSize
			pkt = rcv_pkt(transport2physical_port, interface)
			for p in 0..(pkt.length-1)
				if j >= dataSize
					vestigio = pkt[p..pkt.length]
					print("Vestigio: ", vestigio,"\n\n")
					break
				end
				pacote.print(pkt[p].ord.to_s(2).rjust(10, '0'))
				j += 1
			end


			# Caso seja um pacote ACK
			if pkt[13] != 0
				print("ACK: ", pkt[13])
				break
			else
				# Caso seja um pacote com dados
				if pkt[pkt.length-7..pkt.length-1] == "LASTSEG"
					if vestigio == ""
						transferencia_aberta = false
					end
					break
				end
			end
		end
	end

	pacote.close()
	last_seg = false

	# Ler dados do pacote
	pdu = File.read('pacote.txt')

	# Verificacao se ha colisao
	# colisao = rand(1...100) > 70
	# while colisao
	# 	puts("Ocorreu colisao no envio do pacote "+package_index.to_s+"\n")
	# 	sleep(rand(1...100)/100)
	# 	puts("Reenvio do pacote "+package_index.to_s+"\n")
	# 	colisao = rand(1...100) > 70
	# end

	# Envio da pdu
	tcpConnect(host, port, pdu)
	puts("Envio do quador "+package_index.to_s+" realizado com sucesso\n")
	package_index += 1

end

puts("---------------------------------------")
puts("\n\nEnvio do arquivo HTML para o buffer de entrada do cliente\n\n")