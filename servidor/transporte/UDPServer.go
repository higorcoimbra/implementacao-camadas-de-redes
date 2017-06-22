package main
 
import (
    "fmt"
    "net"
    "os"
    "bytes"
    "strings"
    "strconv"
)

/* A Simple function to verify error */
func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }
}

func makeTransportHeader(source_port string,destination_port string,buffer_size int)(header bytes.Buffer){
    header.WriteString(strings.Split(source_port,":")[1])
    header.WriteString(" ")
    header.WriteString(strings.Split(destination_port,":")[1])
    header.WriteString("\n")
    header.WriteString(strconv.Itoa(buffer_size+len(source_port)+len(destination_port)))
    header.WriteString("\n")
    return header
}


func getSourceDestinationPort(header string)(string,string){
	var destination_port bytes.Buffer
	var source_port bytes.Buffer
	destination_port.WriteString(":")
	destination_port.WriteString(strings.Split(header," ")[1])
	source_port.WriteString(":")
	source_port.WriteString(strings.Split(header," ")[0])
	return source_port.String(),destination_port.String()
}

func rdt_rcv(physical2transport_connection net.Conn)(pkt []byte){
	pkt = make([]byte, 1024)

    fmt.Println("Recebendo pacote da camada física...")
    _,err := physical2transport_connection.Read(pkt)

    CheckError(err)

    return pkt
}


/*
    Funcao responsavel por verificar se o numero de sequencia 
    do pacote recebido e igual ao esperado
*/
func hasSeqNum(rcvpkt string, expectedseqnum int) (bool) {
    //retira do pacote recebido o numero de sequencia
    data := strings.Split(rcvpkt, "\n")
    seqnum := data[1]
    //compara com o numero esperado
    return seqnum == string(expectedseqnum)
}

/*
    Extrai os dados do pacote
*/
func extract(rcvpkt string)(string,string,string){
    dados := strings.Split(rcvpkt,"\n")
    portas := strings.Split(dados[0]," ")
    return portas[0],portas[1],dados[1]
}

/*
    Envia dados para a camada de aplicação
*/
func deliverData(destination_port string, data string){
    destination_address := ":" + destination_port
    
    //enviando conteúdo do pacote para a camada de aplicação
    fmt.Println("Enviando pacote para a camada de aplicação...")
    transport2app_port,err := net.ResolveTCPAddr("tcp",destination_address)
    CheckError(err)
    transport2app_connection, err := net.DialTCP("tcp", nil, transport2app_port)
    CheckError(err)
    _,err = transport2app_connection.Write([]byte(data))
    CheckError(err)
    transport2app_connection.Close()
    fmt.Println("Pacote enviado com sucesso para a aplicação.")
}

/*
    Monta o pacote ACK
*/
func makePkt(expectedseqnum string, source_port string, destination_port string)(pkt string){
    pkt = destination_port + " " + source_port + "\n" + expectedseqnum + "\n" + expectedseqnum + "\n"
    return pkt
}

/*
    Envia pacote a camada física
*/
func udtSend(segment string,transport2physical_address string){
    fmt.Println("Enviando pacote para a camada física...")
    physical_port, err := net.ResolveTCPAddr("tcp", transport2physical_address)
    CheckError(err)
    physical_connection, err := net.DialTCP("tcp", nil, physical_port)
    CheckError(err)
    _, err = physical_connection.Write([]byte(segment))
    CheckError(err)
    physical_connection.Close()
}

func main(){
    physical2transport_address := ":8002"
    transport2physical_address := ":8008"
    buf := make([]byte, 1024)

    var opcao_trasmissao string
    opcao_trasmissao = "tcp"
	
    physical2transport_port,err := net.ResolveTCPAddr("tcp",physical2transport_address)
    CheckError(err)
    physical2transport_listener, err := net.ListenTCP("tcp", physical2transport_port)
    CheckError(err)
    physical2transport_connection, err := physical2transport_listener.Accept() 
    CheckError(err)

    if opcao_trasmissao == "udp" {
    	//lendo dados da camada física
        buffer_size,err := physical2transport_connection.Read(buf)
        CheckError(err)
        fmt.Println(string(buf[0:buffer_size]))
        fmt.Println("Mensagem recebida com sucesso da camada física...")
        physical2transport_connection.Close()

        //enviando conteúdo do pacote para a camada de aplicação
        fmt.Println("Enviando pacote para a camada de aplicação...")
        source_address,destination_address := getSourceDestinationPort(string(buf[0:9]))
        app_content := string(buf[14:])
        transport2app_port,err := net.ResolveTCPAddr("tcp",destination_address)
    	CheckError(err)
        transport2app_connection, err := net.DialTCP("tcp", nil, transport2app_port)
        CheckError(err)
    	buffer_size,err = transport2app_connection.Write([]byte(app_content))
        CheckError(err)
        transport2app_connection.Close()
        fmt.Println("Pacote enviado com sucesso para a aplicação.")

        //recebendo resposta HTTP da camada de aplicação
        fmt.Println("Recebendo resposta HTTP da camada de aplicação...")
        tmp := source_address
        source_address = destination_address
        destination_address = tmp
        app2transport_port,err := net.ResolveTCPAddr("tcp",source_address)
    	CheckError(err)
        app2transport_listener, err := net.ListenTCP("tcp", app2transport_port)
        CheckError(err)
    	appp2transport_connection, err := app2transport_listener.Accept()
    	CheckError(err)
    	buffer_size,err = appp2transport_connection.Read(buf)
    	appp2transport_connection.Close()
        
        //enviando resposta HTTP para a camada física
        //fazendo cabeçalho UDP
    	var pdu_content bytes.Buffer
        application_content := string(buf[0:buffer_size])
        transport_header := makeTransportHeader(source_address,destination_address,buffer_size)
        pdu_content.WriteString(transport_header.String())
        pdu_content.WriteString(application_content)
        transport2physical_port,err := net.ResolveTCPAddr("tcp",transport2physical_address)
    	CheckError(err)
        transport2physical_connection, err := net.DialTCP("tcp", nil, transport2physical_port)
        CheckError(err)
    	buffer_size,err = transport2physical_connection.Write([]byte(pdu_content.String()))
        CheckError(err)
        transport2physical_connection.Close()
    } else if opcao_trasmissao == "tcp" {

        rcvpkt := make([]byte, 1024)
        minSize := 10
        expectedseqnum := 0
        /*
            maquina de estados do destinatario
        */
        for ; ; {
            rcvpkt = rdt_rcv(physical2transport_connection)

            fmt.Println(string(rcvpkt))

            // caso receba um pacote com tamanho 
            if (len(rcvpkt) <= minSize) {
                break;
            }

            if (hasSeqNum(string(rcvpkt), expectedseqnum)) {
                source_port,destination_port,data := extract(string(rcvpkt))
                deliverData(data,destination_port)
                sndpkt := makePkt(string(expectedseqnum), source_port, destination_port)
                udtSend(sndpkt, transport2physical_address)
                expectedseqnum += 1
            }
        }


    }

}