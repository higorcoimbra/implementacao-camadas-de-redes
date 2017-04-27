=begin
	Camada física
	Servidor
=end

require 'socket'
require 'macaddr'

# variaveis de configuracao de host
host = '127.0.0.1'
port = 8000

# variaveis de transmissao
transmissionServer = 100
gargalo = transmissionServer
macServidor = Mac.address

server = TCPServer.open(port)

# conexao para enviar tamanho da transmissao
client = server.accept
client.puts(transmissionServer)
client.close

# conexao para enviar endereco mac
client = server.accept
client.puts(macServidor)

destino = File.new("destino", "w")

# conexao para pegar o pacote
while (1)

	client = server.accept
	pacote = client.read()
	if (pacote.length < 10)
		break
	end

	dados = pacote[100..pacote.length]

	# extrai bytes dos bits
	i = 0
	while (i < dados.length-1)
		byte = ""
		j = 0
		while (j < 10)
			byte += dados[i]
			i += 1
			j += 1
		end
		destino.print(byte.to_i(2).chr)
	end

end
