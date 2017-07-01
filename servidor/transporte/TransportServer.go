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
    seqnum := ""
    if (strings.Index(data[1], "_") != -1) {
        seqnum = strings.Split(data[1], "_")[0]
    } else {
        seqnum = data[1]
    }
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
    print("\n-----data aplication----\n")
    print(data_aplication)
    print("\n-----------\n")
    if(strings.Contains(data_aplication,"LASTSEG")){
        data_aplication = data_aplication[:strings.Index(data_aplication,"LASTSEG")]
    }else{
        data_aplication = data_aplication[:strings.Index(data_aplication,"TRAILER")]
    }

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
    fmt.Println("Construindo ACK: ", expectedseqnum)
    valor,_ := strconv.Atoi(expectedseqnum)
    if (valor < 10) {
        pkt = destination_port + " " + source_port + "\n" + expectedseqnum + "_\n" + expectedseqnum + "_"
    } else {
        pkt = destination_port + " " + source_port + "\n" + expectedseqnum + "\n" + expectedseqnum
    }

    return pkt
}

func convert(str []byte)(converted string){
    
    converted = ""
    for i := 0;i < len(str); i +=1{
        if(str[i] != 00000000){
            converted += string(str[i])
        }
    }
    return converted
}

/*
    Envia pacote a camada física
*/
func ACKSend(segment string,transport2physical_address string){
    fmt.Println("Enviando confirmacao de pacote para a camada física...")
    physical_port, err := net.ResolveTCPAddr("tcp", transport2physical_address)
    CheckError(err)
    physical_connection, err := net.DialTCP("tcp", nil, physical_port)
    for ; ; {
        physical_connection, err = net.DialTCP("tcp", nil, physical_port)
        if (err == nil) {
            break
        }
    }    
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

    physical2transport_address := ":8006"
    transport2app_address := ":8007"
    transport2physical_address := ":8010"
    
    physical2transport_port,err := net.ResolveTCPAddr("tcp",physical2transport_address)
    CheckError(err)
    physical2transport_listener, err := net.ListenTCP("tcp", physical2transport_port)
    CheckError(err)

    rcvpkt := make([]byte, 1024)
    expectedseqnum := 1
    app_content := []AppContent{}
    transport2app_content := ""

    /*
        maquina de estados do destinatario
    */
    for {

        /*
            Recebendo segmento da camada fisica
        */
        print("Recebendo segmento da camada física...")
        physical2transport_connection, err := physical2transport_listener.Accept() 
        CheckError(err)
        _,err = physical2transport_connection.Read(rcvpkt)
        rcvpkt_formated := convert(rcvpkt)
        CheckError(err)
        _,_,_,data := extract(rcvpkt_formated)
        app_content = append(app_content, AppContent{expectedseqnum, data})
        if (matchSeqNum(string(rcvpkt_formated), expectedseqnum)) {
            print(string(rcvpkt_formated))
            print(string(rcvpkt_formated[26:33]))
            source_port,destination_port,_,data := extract(string(rcvpkt))
            app_content = append(app_content, AppContent{expectedseqnum, data})
            sndACK := makeACK(strconv.Itoa(expectedseqnum), source_port, destination_port)
            ACKSend(sndACK, transport2physical_address)
            expectedseqnum += 1
        } else {
            source_port,destination_port,_,_ := extract(string(rcvpkt_formated))
            sndACK := makeACK(strconv.Itoa(expectedseqnum-1), source_port, destination_port)
            ACKSend(sndACK, transport2physical_address)
        }

        if(strings.Contains(rcvpkt_formated,"LASTSEG")){
            index := strings.Index(rcvpkt_formated,"LASTSEG")
            print(rcvpkt_formated[index:index+len("LASTSEG")])
            if (rcvpkt_formated[index:index+len("LASTSEG")] == "LASTSEG") {
                physical2transport_connection.Close()
                break;
            }
        }
    }

    sort.Sort(BySequenceNumber(app_content))
    for i := 0; i < len(app_content); i++ {
        transport2app_content += app_content[i].Dado
    }
    //deliverData(data,destination_port) <---------- enviar mensagem completa pra camada de aplicacao do servidor
    print("\n\n\nCONTEUDO\n\n")
    print(transport2app_content)


    //enviando conteúdo para a camada de aplicação
    print("Enviando conteúdo para a camada de aplicação...")
    transport2app_port,err := net.ResolveTCPAddr("tcp",transport2app_address)
    CheckError(err)
    transport2app_connection, err := net.DialTCP("tcp", nil, transport2app_port)
    CheckError(err)
    _,err = transport2app_connection.Write([]byte(transport2app_content))
    CheckError(err)
    transport2app_connection.Close()
    print("Pacote enviado com sucesso para a aplicação.")

}
