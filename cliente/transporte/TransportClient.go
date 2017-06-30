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

func makeTransportHeaderUDP(source_port string,destination_port string,sequence_number int)(header bytes.Buffer){
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
    physical_port, err := net.ResolveTCPAddr("tcp", transport2network_address)
    CheckError(err)
    physical_connection, err := net.DialTCP("tcp", nil, physical_port)
    for{
        if err == nil{
           break 
        }
        physical_connection, err = net.DialTCP("tcp", nil, physical_port)
    }
    CheckError(err)
    _, err = physical_connection.Write([]byte(segment))
    //CheckError(err)
    physical_connection.Close()
}


func main() {
	print := fmt.Println
    
    app2transport_address := ":8001"
    transport2network_address := ":8002"
    source_port := app2transport_address
    destination_port := ":8006"             // porta em que o processo esta rodando no servidor

    WINDOW_SIZE := 3
    DATA_SIZE := 9
    nextseqnum := 1
    timeSleep := time.Nanosecond * 400000000
    base := 1
    string_slice := make([]byte,0)
    rdt_buffer := make([]string, 0)

    var data string
    var segment string
    var header bytes.Buffer
    
    // Conexao com a camada superior
    app2transport_port,err := net.ResolveUDPAddr("udp",app2transport_address)
    CheckError(err)
    app2transport_connection, err := net.ListenUDP("udp", app2transport_port)
    CheckError(err)

    // Le dados da camada superior
    buf := make([]byte, 1024)
    buffer_size,_,err := app2transport_connection.ReadFromUDP(buf)
    CheckError(err)
    application_content := string(buf[0:buffer_size])

    // Controi buffer com os dados recebidos da aplicacao
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

    print("\n***************** CONTEUDO RECEBIDO DA APLICACAO ******************\n")
    print(application_content)

    /*
        Maquina de estados do remetente
    */
    for ; len(rdt_buffer) != 0 ; {
        print("\n******** CONTROI SEGMENTO E ENVIA PARA A CAMADA DE REDES ********")
        
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
            header = makeTransportHeaderUDP(source_port,destination_port,nextseqnum)
            segment = makeSegment(header.String(), data, len(rdt_buffer))
            print("\n------------------------- Segmento --------------------------------\n")
            print(segment)
            print("\n---------------------------------------------------------------------")
            udt_send(segment,transport2network_address)
            nextseqnum = nextseqnum + 1
            base += 1
        }

        time.Sleep(timeSleep)

        print("\n***************************** OK *******************************\n")
    }   
}