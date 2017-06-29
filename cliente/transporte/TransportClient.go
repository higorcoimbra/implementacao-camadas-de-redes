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
        print("Error: " , err)
        os.Exit(0)
    }
}
 
func makeTransportHeaderUDP(source_port string,destination_port string,buffer_size int)(header bytes.Buffer){
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
    if (sequence_number < 10) {
        header.WriteString(strconv.Itoa(sequence_number)+"_")
    } else {
        header.WriteString(strconv.Itoa(sequence_number))
    }
    header.WriteString("\n")
    header.WriteString("0_")
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

func makeSegment(header string, data string, len_rdt_buffer int)(string){
    var segment bytes.Buffer
    segment.WriteString(header)
    segment.WriteString(data)
    if len_rdt_buffer != 0 {
        segment.WriteString("TRAILER")
    }else{
        segment.WriteString("LASTSEG")
    }
    return segment.String()
}

func udt_send(segment string,transport2network_address string){
    print("\nEnviando pacote para a camada de rede...")
    print(transport2network_address)
    physical_port, err := net.ResolveTCPAddr("tcp", transport2network_address)
    CheckError(err)
    print("\nDialTCP\n")
    physical_connection, err := net.DialTCP("tcp", nil, physical_port)
    for{
        if err == nil{
           break 
        }
        physical_connection, err = net.DialTCP("tcp", nil, physical_port)
    }
    CheckError(err)
    print("\nWrite\n")
    _, err = physical_connection.Write([]byte(segment))
    //CheckError(err)
    physical_connection.Close()
}



func printBuffer(buffer []string){
    print("\nRDT_BUFFER:")
    for i := 0; i < 5; i++ {
        print(buffer[i])
    }
    print("END OF BUFFER\n")
}


func main() {
	print := fmt.Println

    var opcao_trasmissao string
    opcao_trasmissao = "tcp"

    var transport_header bytes.Buffer
    var pdu_content bytes.Buffer
    app2transport_address := ":8001"
    transport2network_address := ":8002"
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
        //print("")
        //print("")
        transport_header = makeTransportHeaderUDP(source_port,destination_port,buffer_size)
        pdu_content.WriteString(transport_header.String())
        pdu_content.WriteString(application_content)
        pdu_content.WriteString("TRAILER")
        //print(pdu_content.String())

        //enviando pacote a camada de rede
        print("Enviando pacote para a camada de rede...")
        physical_port, err := net.ResolveTCPAddr("tcp", transport2network_address)
        CheckError(err)
        physical_connection, err := net.DialTCP("tcp", nil, physical_port)
        CheckError(err)
        _, err = physical_connection.Write([]byte(pdu_content.String()))
        CheckError(err)
        physical_connection.Close()

        //recebendo pacote da física pra passar pra aplicação
        physical_content := make([]byte,1024)
        print("Recebendo pacote da camada física...")
        physical2transport_port,err := net.ResolveTCPAddr("tcp",physical2transport_address)
        CheckError(err)
        physical2transport_listener, err := net.ListenTCP("tcp", physical2transport_port)
        CheckError(err)
        physical2transport_connection, err := physical2transport_listener.Accept() 
        CheckError(err)
        size,err := physical2transport_connection.Read(physical_content)
        CheckError(err)
        print(string(physical_content[0:size]))
        print("Pacote recebido com sucesso!")

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
        WINDOW_SIZE := 3
        DATA_SIZE := 9
        nextseqnum := 1
        timeSleep := time.Nanosecond * 400000000
        base := 1
        string_slice := make([]byte,0)
        rdt_buffer := make([]string, 0)
        //timeout := int64(1000000000)
        //stop_timer := true

        //var ack int
        var data string
        var segment string
        var header bytes.Buffer

        //var current int64
        //var total_interval int64
        //var start_timer int64
        //var interval int64

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

        window_buffer := make([]string, 0)
        /*
        	Maquina de estados para o remetente
        */
        
        print("---APPLICATION CONTENT----")
        print(application_content)

    
        //print("\nAbrindo conexao para receber ACK...")
        //physical2transport_port, _ := net.ResolveTCPAddr("tcp",physical2transport_address)
        //physical2transport_listener, _ := net.ListenTCP("tcp", physical2transport_port)

        print("\ntamanho rdt buffer\n")
        print(len(rdt_buffer))


        //physical_connection, _ := net.Dial("tcp", "127.0.0.1"+transport2network_address)
        for ; len(rdt_buffer) != 0 ; {
            if(len(rdt_buffer) > 0){
                data = rdt_buffer[0]
            }

            /*
                Envio de dados caso o proximo pacote esteja dentro da janela
            */

            if(nextseqnum < base+WINDOW_SIZE){
                if (len(rdt_buffer) > 0) {
                    rdt_buffer = rdt_buffer[1:]
                }
                header = makeTransportHeaderTCP(source_port,destination_port,nextseqnum)
                segment = makeSegment(header.String(), data, len(rdt_buffer))
                print("antes do Fprintf\n")
                print(segment,"\n")
                udt_send(segment,transport2network_address)
                //fmt.Fprintf(physical_connection, segment)
                print("depois do Fprintf\n")
                window_buffer = append(window_buffer,segment)
                // if(base == nextseqnum){
                // 	stop_timer = false
                //     start_timer = time.Now().UnixNano()
                //     total_interval = 0
                // }
                nextseqnum = nextseqnum + 1
                base += 1
            }

            /*
				Verificacao do timeout do pacote base
            */
            /*
			if (stop_timer == true) {
				current = 0
			} else {
            	current = time.Now().UnixNano()
            }
            //print(ack," ",current," ",start_timer," ",current-start_timer," ",total_interval)

            if(current - start_timer -total_interval > timeout) {
                print("Tempo excedido")
                //start_timer = time.Now().UnixNano()
                //total_interval = 0
                print("\n", nextseqnum)
                // physical_port, err := net.ResolveTCPAddr("tcp", transport2network_address)
                // CheckError(err)
                // physical_connection, err := net.DialTCP("tcp", nil, physical_port)
                // CheckError(err)
                for i := 0; i < nextseqnum-1; i++ {
                    //_, err = physical_connection.Write([]byte(window_buffer[i]))
                    //CheckError(err)
                    print(window_buffer[i])
                    udt_send(window_buffer[i], transport2network_address)
                }
                //physical_connection.Close()    
            }*/

            /*
				Verificacao de recepcao de ack do destino
            */
            /*{
                read_timeout := 1e9*time.Nanosecond
                physical_content := make([]byte,1024)
                inicio := time.Now().UnixNano()
                ack = 0
                print("\nSetDeadline...")
                physical2transport_listener.SetDeadline(time.Now().Add(read_timeout))
                print("\nAccept...")
                physical2transport_connection, err := physical2transport_listener.Accept()
                print("\nSetReadDeadline...\n")
                if err == nil {
                    print("\nRecebeu ACK...")
                    physical2transport_connection.SetReadDeadline(time.Now().Add(read_timeout))    
                    print("Read...\n")
                    _,_ = physical2transport_connection.Read(physical_content)
                    print("\nRecebeu ack - conteudo dele\n")
                    print(string(physical_content))
                    ack = int(physical_content[12])
                }

                time.Sleep(400000000)
                interval = time.Now().UnixNano() - inicio


                //interval, ack = rdt_rcv(physical2transport_listener)
                total_interval = total_interval + interval
                if ack != 0 {
                	base = ack+1
                	if base == nextseqnum {
                		stop_timer = true
                	} else {
                		start_timer = time.Now().UnixNano()
                	}
                }
            }*/
            time.Sleep(timeSleep)
        }   
    }
}