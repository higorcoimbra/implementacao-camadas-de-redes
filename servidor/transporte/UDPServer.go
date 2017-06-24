package main
 
import (
    "fmt"
    "net"
    "os"
    "bytes"
    "strings"
    "strconv"
    "sort"
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


/*
    Funcao responsavel por verificar se o numero de sequencia 
    do pacote recebido e igual ao esperado
*/
func matchSeqNum(rcvpkt string, expectedseqnum int) (bool) {
    //retira do pacote recebido o numero de sequencia
    data := strings.Split(rcvpkt, "\n")
    seqnum := data[1]

    //compara com o numero esperado
    return seqnum == strconv.Itoa(expectedseqnum)
}

/*
    Extrai os dados do pacote
*/
func extract(rcvpkt string)(string,string,string,string){
    rcvpkt_formated := fmt.Sprintf(rcvpkt)
    dados := strings.Split(rcvpkt_formated,"\n")
    portas := strings.Split(dados[0]," ")
    
    data_aplication := ""
    for i := 3; i < len(dados)-1; i++ {
        data_aplication += dados[i] + "\n"
    }  

    data_aplication = data_aplication[:len(data_aplication)-8]

    return portas[0],portas[1],dados[1],data_aplication
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
func makeACK(expectedseqnum string, source_port string, destination_port string)(pkt string){
    pkt = destination_port + " " + source_port + "\n" + expectedseqnum + "\n" + expectedseqnum
    return pkt
}

/*
    Envia pacote a camada física
*/
func ACKSend(segment string,transport2physical_address string){
    fmt.Println("Enviando confirmacao de pacote para a camada física...")
    physical_port, err := net.ResolveTCPAddr("tcp", transport2physical_address)
    CheckError(err)
    physical_connection, err := net.DialTCP("tcp", nil, physical_port)
    CheckError(err)
    _, err = physical_connection.Write([]byte(segment))
    CheckError(err)
    physical_connection.Close()
}


type AppContent struct {
    SequenceNumber  int
    Dado string
}

type BySequenceNumber []AppContent

func (a BySequenceNumber) Len() int           { return len(a) }
func (a BySequenceNumber) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySequenceNumber) Less(i, j int) bool { return a[i].SequenceNumber < a[j].SequenceNumber }

func main(){
    print := fmt.Println

    physical2transport_address := ":8002"
    transport2physical_address := ":8008"
    buf := make([]byte, 1024)

    var opcao_trasmissao string
    opcao_trasmissao = "tcp"
	
    physical2transport_port,err := net.ResolveTCPAddr("tcp",physical2transport_address)
    CheckError(err)
    physical2transport_listener, err := net.ListenTCP("tcp", physical2transport_port)
    CheckError(err)

    if opcao_trasmissao == "udp" {
    	//lendo dados da camada física
        physical2transport_connection, err := physical2transport_listener.Accept() 
        CheckError(err)
        buffer_size,err := physical2transport_connection.Read(buf)
        CheckError(err)
        print(string(buf[0:buffer_size]))
        print("Mensagem recebida com sucesso da camada física...")
        physical2transport_connection.Close()

        //enviando conteúdo do pacote para a camada de aplicação
        print("Enviando pacote para a camada de aplicação...")
        source_address,destination_address := getSourceDestinationPort(string(buf[0:9]))
        app_content := string(buf[14:])
        transport2app_port,err := net.ResolveTCPAddr("tcp",destination_address)
    	CheckError(err)
        transport2app_connection, err := net.DialTCP("tcp", nil, transport2app_port)
        CheckError(err)
    	buffer_size,err = transport2app_connection.Write([]byte(app_content))
        CheckError(err)
        transport2app_connection.Close()
        print("Pacote enviado com sucesso para a aplicação.")

        //recebendo resposta HTTP da camada de aplicação
        print("Recebendo resposta HTTP da camada de aplicação...")
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
        expectedseqnum := 1
        app_content := []AppContent{}
        transport2app_content := ""

        /*
            maquina de estados do destinatario
        */
        for ; ; {

            /*
                Recebendo segmento da camada fisica
            */

            print("Recebendo segmento da camada física...")
            physical2transport_connection, err := physical2transport_listener.Accept() 
            CheckError(err)
            _,err = physical2transport_connection.Read(rcvpkt)
            CheckError(err)

            if (matchSeqNum(string(rcvpkt), expectedseqnum)) {
                source_port,destination_port,_,data := extract(string(rcvpkt))
                app_content = append(app_content, AppContent{expectedseqnum, data})
                sndACK := makeACK(strconv.Itoa(expectedseqnum), source_port, destination_port)
                ACKSend(sndACK, transport2physical_address)
                expectedseqnum += 1
            }

            if (string(rcvpkt[24:31]) == "LASTSEG") {
                physical2transport_connection.Close()
                break;
            }
        }

        sort.Sort(BySequenceNumber(app_content))
        for i := 0; i < len(app_content); i++ {
            transport2app_content += app_content[i].Dado
        }
        //deliverData(data,destination_port) <---------- enviar mensagem completa pra camada de aplicacao do servidor
        print("\n\n\nCONTEUDO\n\n")
        print(transport2app_content)

    }

}
