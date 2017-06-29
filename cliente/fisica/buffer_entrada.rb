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
port = 8004
physical2transport_port = 8009

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
pacote = ""
final = "*******"
transferencia_aberta = true

while (transferencia_aberta)

	print("Aguardando quadro do servidor ...\n\n")
	client = server.accept
	quadro = client.read()
	client.close

	# Pega apenas os dados retirando o cabecalho
	pacotes = quadro[164..quadro.length]
	puts("Quadro "+package_index.to_s+" recebido com sucesso\n\n")
	package_index += 1

	# Extrai bytes dos bits
	i = 0
	qtdByte = 1
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

		# Checa se o pacote e' um ack                              \\??
		if qtdByte == 14
			if byte.to_i(2).chr != "0"
				destino.close()
				tcpConnect(host,physical2transport_port,File.read("destino"))
				destino = File.new("destino", "w")
			end
		end
		qtdByte += 1

		if final == "TRAILER" or final == "LASTSEG"
			destino.close()
			puts(File.read("destino"))
			tcpConnect(host,physical2transport_port,File.read("destino"))
			if final == "LASTSEG"
				transferencia_aberta = false
			else
				destino = File.new("destino", "w")
			end
		end
	end

end

puts("Envio do arquivo HTML para camada de transporte do cliente\n\n")


