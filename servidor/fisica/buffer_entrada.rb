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

def pullAndDrag(final, byte)
	for i in 0..final.length-2
		final[i]  = final[i+1]
	end
	final[final.length-1] = byte
	return final
end
# Variaveis de configuracao de host
host = '127.0.0.1'
port = 8000
physical2transport_port = 8002

# Variaveis de transmissao
transmissionServer = 120
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
pacote = ""
final = "*******"

while (1)

	print("Aguardando quadro do cliente")
	client = server.accept
	quadro = client.read()
	client.close

	# Encerramento da transferencia
	if (quadro.length < 10)
		break
	end

	# Pega apenas os dados retirando o cabecalho
	pacotes = quadro[164..quadro.length]


	# Extrai bytes dos bits
	i = 0
	while (i < pacotes.length-1)
		byte = ""
		j = 0
		while (j < 10)
			byte += pacotes[i]
			i += 1
			j += 1
		end
		final = pullAndDrag(final, byte.to_i(2).chr)
		destino.print(byte.to_i(2).chr)
		print("<",final,">\n")
		if final == "TRAILER"
			tcpConnect(host,physical2transport_port,File.read("destino"))
			puts("Envio da mensagem HTTP para camada de transporte do servidor\n\n")
		end
	end

	puts("Quadro "+package_index.to_s+" recebido com sucesso\n")
	package_index += 1

end

destino.close()
puts("\n\nMensagem HTTP recebida com sucesso do buffer de saida do cliente\n\n")


