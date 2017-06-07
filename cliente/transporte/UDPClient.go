package main
 
import (
    "fmt"
    "net"
    "os"
    "bytes"
    "strconv"
    "strings"
    "time"
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

func makeTransportHeaderTCP(source_port string,destination_port string,sequence_number int)(header bytes.Buffer){
    header.WriteString(strings.Split(source_port,":")[1])
    header.WriteString(" ")
    header.WriteString(strings.Split(destination_port,":")[1])
    header.WriteString("\n")
    header.WriteString(string(sequence_number))
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

func makeSegment(header string, data string)(string){
    var segment bytes.Buffer
    segment.WriteString(header)
    segment.WriteString(data)
    return segment.String() 
}

func udt_send(segment string,transport2physical_address string){
    fmt.Println("Enviando pacote para a camada física...")
    physical_port, err := net.ResolveTCPAddr("tcp", transport2physical_address)
    CheckError(err)
    physical_connection, err := net.DialTCP("tcp", nil, physical_port)
    CheckError(err)
    _, err = physical_connection.Write([]byte(segment))
    CheckError(err)
    physical_connection.Close()
}

func rdt_rcv(physical2transport_address string)(intervalo int, pacote string){
    read_timeout := time.Nanosecond
    physical_content := make([]byte,1024)
    inicio := time.Now().Nanosecond()
    
    fmt.Println("Recebendo ACK da camada física...")
    physical2transport_port,err := net.ResolveTCPAddr("tcp",physical2transport_address)
    CheckError(err)
    physical2transport_listener, err := net.ListenTCP("tcp", physical2transport_port)
    CheckError(err)
    physical2transport_connection, err := physical2transport_listener.Accept() 
    CheckError(err)
    physical2transport_connection.SetReadDeadline(time.Now().Add(read_timeout))
    _,err = physical2transport_connection.Read(physical_content)
    CheckError(err)

    intervalo = time.Now().Nanosecond() - inicio
    return intervalo, string(physical_content)
}

func printBuffer(buffer []string){
    fmt.Println("\nRDT_BUFFER:")
    for i := 0; i < 5; i++ {
        fmt.Println(buffer[i])
    }
    fmt.Println("END OF BUFFER\n")
}


func main() {

    var opcao_trasmissao string
    opcao_trasmissao = "tcp"

    var transport_header bytes.Buffer
    var pdu_content bytes.Buffer
    app2transport_address := ":8001"
    transport2physical_address := ":8006"
    physical2transport_address := ":8009"
    source_port := app2transport_address
    destination_port := ":8007"
    
    app2transport_port,err := net.ResolveUDPAddr("udp",app2transport_address)
    CheckError(err)
 
    app2transport_connection, err := net.ListenUDP("udp", app2transport_port)
    CheckError(err)
    defer app2transport_connection.Close()
 
    buf := make([]byte, 1024)
    buffer_size,_,err := app2transport_connection.ReadFromUDP(buf)
    CheckError(err)
    application_content := string(buf[0:buffer_size])

    if opcao_trasmissao == "udp" {

        //formando pacote na camada de transporte
        //fmt.Println("")
        //fmt.Println("")
        transport_header = makeTransportHeader(source_port,destination_port,buffer_size)
        pdu_content.WriteString(transport_header.String())
        pdu_content.WriteString(application_content)
        //fmt.Println(pdu_content.String())

        //enviando pacote a camada física
        fmt.Println("Enviando pacote para a camada física...")
        physical_port, err := net.ResolveTCPAddr("tcp", transport2physical_address)
        CheckError(err)
        physical_connection, err := net.DialTCP("tcp", nil, physical_port)
        CheckError(err)
        _, err = physical_connection.Write([]byte(pdu_content.String()))
        CheckError(err)
        physical_connection.Close()

        //recebendo pacote da física pra passar pra aplicação
        physical_content := make([]byte,1024)
        fmt.Println("Recebendo pacote da camada física...")
        physical2transport_port,err := net.ResolveTCPAddr("tcp",physical2transport_address)
        CheckError(err)
        physical2transport_listener, err := net.ListenTCP("tcp", physical2transport_port)
        CheckError(err)
        physical2transport_connection, err := physical2transport_listener.Accept() 
        CheckError(err)
        size,err := physical2transport_connection.Read(physical_content)
        CheckError(err)
        fmt.Println(string(physical_content[0:size]))
        fmt.Println("Pacote recebido com sucesso!")

        //enviando resposta HTTP pra camada de aplicação
        app_content := string(physical_content[14:])
        _,destination_address := getSourceDestinationPort(string(physical_content[0:9]))
        transport2app_port,err := net.ResolveTCPAddr("tcp",destination_address)
        CheckError(err)
        transport2app_connection, err := net.DialTCP("tcp", nil, transport2app_port)
        CheckError(err)
        buffer_size,err = transport2app_connection.Write([]byte(app_content))
        CheckError(err)
        transport2app_connection.Close()

    } else if opcao_trasmissao == "tcp" {
        WINDOW_SIZE := 8
        DATA_SIZE := 9
        nextseqnum := 1
        timeSleep := time.Nanosecond * 400000000
        base := 1
        string_slice := make([]byte,0)
        rdt_buffer := make([]string, 0)
        timeout := 100000000000

        //var ack string
        //var interval int
        var data string
        var segment string
        var header bytes.Buffer
        var total_interval int
        var start_timer int
        var current int
        //var stop_timer time.Time
        j := 0
        for i := 0; i < len(application_content); i++ {
            string_slice = append(string_slice,application_content[i])
            if(j == DATA_SIZE || i == len(application_content)-1){
                rdt_buffer = append(rdt_buffer,string(string_slice))
                string_slice = string_slice[:0]
                j = -1            
            }  
            j = j+1
        }

        window_buffer := make([]string,WINDOW_SIZE)
        //loop principal da máquina de estados

        for ; len(rdt_buffer) > 30; {
            data = rdt_buffer[0]

            // envia pacote caso o numero de sequencia esteja dentro da janela
            if(nextseqnum < base+WINDOW_SIZE){
                rdt_buffer = rdt_buffer[1:]
                header = makeTransportHeaderTCP(source_port,destination_port,nextseqnum)
                segment = makeSegment(header.String(),data)
                //udt_send(segment,transport2physical_address)
                window_buffer = append(window_buffer,segment)
                if(base == nextseqnum){
                    start_timer = time.Now().Nanosecond()*0
                    total_interval = time.Now().Nanosecond()*0
                }
                nextseqnum = nextseqnum + 1
            }

            // verifica se o timeout do pacote base foi excedido
            current = time.Now().Nanosecond()
            current = current - total_interval
            if(current - start_timer > timeout) {
                fmt.Println("Tempo excedido")
                fmt.Println(window_buffer[0])
                start_timer = time.Now().Nanosecond()*0
                total_interval = time.Now().Nanosecond()*0
                // for i := 0; i < nextseqnum; i++ {
                //     udt_send(window_buffer[i],transport2physical_address)
                // }
            }

            fmt.Println(current)

            time.Sleep(timeSleep)

            // verifica se chegou um ack
            //interval, _ = rdt_rcv(physical2transport_address)
            //total_interval = total_interval + interval
        }   
    }
}