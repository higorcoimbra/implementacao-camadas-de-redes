package main
 
import (
    "fmt"
    "net"
    "os"
    "bytes"
    "strconv"
    "strings"
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

func main() {
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
 
    //formando pacote na camada de transporte
    buffer_size,_,err := app2transport_connection.ReadFromUDP(buf)
    CheckError(err)
    application_content := string(buf[0:buffer_size])
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
}