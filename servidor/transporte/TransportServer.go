package main
 
import (
    "fmt"
    "net"
    "os"
    "strings"
)

/* A Simple function to verify error */
func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }
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
    if(strings.Contains(data_aplication,"LASTSEG")){
        data_aplication = data_aplication[:strings.Index(data_aplication,"LASTSEG")]
    }else{
        data_aplication = data_aplication[:strings.Index(data_aplication,"TRAILER")]
    }

    return portas[0],portas[1],dados[1],data_aplication
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

func main(){
    print := fmt.Println

    // Configuração de enderecamento para os sockets
    bottomlayer_address := ":8006"
    uplayer_address := ":8007"
    
    // Cria conexão para ouvir a camada inferior
    network2transport_port,err := net.ResolveTCPAddr("tcp", bottomlayer_address)
    CheckError(err)
    // Cria conexão para ouvir a camada superior
    network2transport_listener, err := net.ListenTCP("tcp", network2transport_port)
    CheckError(err)

    // Cria Buffer para receber os segmentos 
    rcvpkt := make([]byte, 1024)
    transport2app_content := ""

    /*
        Maquina de estados do destinatario
    */
    for { 
        /*
            Aguarda Segmento da camada de rede
        */
        print("\n******** AGUARDANDO SEGMENTO DA CAMADA DE REDE ********\n")
        network2transport_connection, err := network2transport_listener.Accept() 
        CheckError(err)
        _,err = network2transport_connection.Read(rcvpkt)
        rcvpkt_formated := convert(rcvpkt)
        CheckError(err)
        print("---------------------- Segmento -----------------------\n")
        print(rcvpkt_formated)
        /* 
            Extrai os dados de aplicacao
        */
        _,_,_,data := extract(rcvpkt_formated)
        print("-------------- Dados extraido do segmento ---------------\n")
        print(data)
        print("---------------------------------------------------------")
        transport2app_content += data

        print("\n************************ OK ***************************\n")
        // Verifica se o segmento recebido foi o ultimo
        if(strings.Contains(rcvpkt_formated,"LASTSEG")){
            index := strings.Index(rcvpkt_formated,"LASTSEG")
            print(rcvpkt_formated[index:index+len("LASTSEG")])
            if (rcvpkt_formated[index:index+len("LASTSEG")] == "LASTSEG") {
                network2transport_connection.Close()
                break;
            }
        }
    }

    print("\n--------------------- CONTEUDO RECEBIDO -----------------------\n\n")
    print(transport2app_content)
    print("--------------------------------------------------------------\n")

    //enviando conteúdo para a camada de aplicação
    print("ENVIA CONTEUDO PARA A CAMADA DE APLICACAO\n")
    transport2app_port,err := net.ResolveTCPAddr("tcp",uplayer_address)
    CheckError(err)
    transport2app_connection, err := net.DialTCP("tcp", nil, transport2app_port)
    CheckError(err)
    _,err = transport2app_connection.Write([]byte(transport2app_content))
    CheckError(err)
    transport2app_connection.Close()
    print("DADOS ENVIADOS.\n")

}
