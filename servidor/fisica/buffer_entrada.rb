=begin
	Camada física
	Servidor
=end

require 'socket'
require 'macaddr'

# Conexao TCP e envio de mensagem
def tcpConnect(host, port, mensagem)
	server = TCPSocket.open(host, port)
	server.puts(mensagem)
	server.close()
end

# Variaveis de configuracao de host
host = '127.0.0.1'
port = 8000
physical2app_port = 8002

# Variaveis de transmissao
transmissionServer = 100
gargalo = transmissionServer
macServidor = Mac.address

server = TCPServer.open(port)

# Conexao para enviar tamanho da transmissao
client = server.accept
client.puts(transmissionServer)
client.close

# Conexao para enviar endereco mac
client = server.accept
client.puts(macServidor)
client.close

# Cria arquivo de destino 
destino = File.new("destino", "w")

# Conexao para pegar o pacote
while (1)

	client = server.accept
	pacote = client.read()
	client.close

	# Encerramento da transferencia
	if (pacote.length < 10)
		break
	end

	# Pega apenas os dados retirando o cabecalho
	dados = pacote[100..pacote.length]

	# Extrai bytes dos bits
	i = 0
	while (i < dados.length-1)
		byte = ""
		j = 0
		while (j < 10)
			byte += dados[i]
			i += 1
			j += 1
		end
		# Escreve no arquivo de destino
		destino.print(byte.to_i(2).chr)
	end

end

destino.close()
tcpConnect(host,physical2app_port,File.read("destino"))




