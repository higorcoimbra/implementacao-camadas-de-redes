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
port = 8004
physical2app_port = 8005

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
package_index = 1
puts("\n\n")
while (1)

	client = server.accept
	pacote = client.read()
	client.close

	# Encerramento da transferencia
	if (pacote.length < 10)
		break
	end

	# Pega apenas os dados retirando o cabecalho
	dados = pacote[164..pacote.length]

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
	puts("Pacote "+package_index.to_s+" recebido com sucesso\n")
	package_index += 1

end

destino.close()
puts("\n\nArquivo HTML recebido com sucesso do buffer de saida do servidor\n\n")

tcpConnect(host,physical2app_port,File.read("destino"))
puts("Envio do arquivo HTML para camada de aplicacao do cliente\n\n")