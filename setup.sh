go get -u "github.com/gliderlabs/ssh"
go get -u "golang.org/x/crypto/ssh"
go get -u "golang.org/x/crypto/ssh/terminal"
go get -u "github.com/integrii/flaggy"
ssh-keygen -f ./id_rsa
go build -o server main.go