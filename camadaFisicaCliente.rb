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


# Variaveis uteis
infinito = 0x3f3f3f3f
headerSize = 14

# Variaveis de configuracao de host
host = '127.0.0.1'
port = 8000
macClient = Mac.addr
macServer = 'aa:aa:aa:aa:aa:aa'

# Variaveis de configuracao da transmissao
transmissionClient = 100
transmissionServer = infinito
gargalo = transmissionClient

# Cria uma conexao TCP para pegar maximo de transmissao
server = TCPSocket.open(host, port)
transmissionServer = Integer(server.gets)
server.close()
gargalo = transmissionClient < transmissionServer ? transmissionClient : transmissionServer

# Verifica se existe espaco para os dados
dataSize = gargalo - headerSize
if dataSize < 1
	puts("Largura de Banda Insuficiente. ")
	tcpConnect(host, port, "acabou")
	exit 1
end

# Pega informacoes de host do servidor
server = TCPSocket.open(host, port)
macServer = server.gets
server.close()

# Le o arquivo que se quer transferir
file = open('exemplo.pdf', "rb")

# Envio dos pacotes - Transformacao em binario, formatacao e envio
# [ MAC Destino - MAC Origem - Ether type | Dados ]
macClientBit = getMacBit(macClient)
macServerBit = getMacBit(macServer)
etherType = "0000"
ends = false

while not ends
	
	# Cria pacote
	# - Escreve cabecalho
	pacote = File.new("pacote.txt", "w")
	pacote.print(macClientBit)
	pacote.print(macServerBit)
	pacote.print(etherType)
	
	# - Escreve dados do pacote
	for i in 0..dataSize
		part = file.read(1)
		if part == nil
			ends = true
			break
		end
		pacote.print(part.ord.to_s(2).rjust(10, '0'))
	end
	pacote.close()
	
	# Ler dados do pacote
	pdu = File.read('pacote.txt')

	# Verificacao se ha colisao
	colisao = rand(1...100) > 90
	while colisao
		sleep(rand(1...100)/100)
		colisao = rand(1...100) > 90
	end

	# Envio da pdu
	tcpConnect(host, port, pdu)

end

# Encerramento da transferencia
tcpConnect(host, port, "acabou")
