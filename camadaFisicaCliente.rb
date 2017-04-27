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


# transforma o mac address em binario
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


# variaveis uteis
infinito = 0x3f3f3f3f

# variaveis de configuracao de host
host = '127.0.0.1'
port = 8000
macClient = Mac.addr
macServer = 'aa:aa:aa:aa:aa:aa'

# variaveis de configuracao da transmissao
transmissionClient = 100
transmissionServer = infinito
gargalo = transmissionClient

# cria uma conexao TCP para pegar maximo de transmissao
server = TCPSocket.open(host, port)
transmissionServer = Integer(server.gets)
gargalo = transmissionClient < transmissionServer ? transmissionClient : transmissionServer

# pega informacoes de host do servidor
server = TCPSocket.open(host, port)
macServer = server.gets

# le o arquivo que se quer transferir
file = open('exemplo.pdf', "rb")

# Envio dos pacotes - Transformacao em binario, formatacao e envio
# [ MAC Destino - MAC Origem - Ether type | Dados | Checksum ]
dataSize = gargalo - 14
if dataSize < 1
	puts("Largura de Banda Insuficiente. ")
	tcpConnect(host, port, "acabou")
	exit 1
end
macClientBit = getMacBit(macClient)
macServerBit = getMacBit(macServer)
etherType = "0000"
ends = false
while not ends
	
	# cria pacote
	pacote = File.new("pacote.txt", "w")
	pacote.print(macClientBit)
	pacote.print(macServerBit)
	pacote.print(etherType)
	
	for i in 0..dataSize
		part = file.read(1)
		if part == nil
			ends = true
			break
		end
		pacote.print(part.ord.to_s(2).rjust(10, '0'))
	end
	pacote.close()
	
	# ler dados do pacote
	pdu = File.read('pacote.txt')
	# tenta enviar, se nao conseguir espera um tempo aleatorio e reenvia
	colisao = rand(1...100) > 90
	while colisao
		sleep(rand(1...100)/100)
		colisao = rand(1...100) > 90
	end

	tcpConnect(host, port, pdu)

end

tcpConnect(host, port, "acabou")
