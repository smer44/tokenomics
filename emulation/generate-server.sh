REPO="quay.io"    #<- or "ghcr.io"
alias swagger='docker run --rm -it  --user $(id -u):$(id -g) -v $HOME:$HOME -w $PWD $REPO/goswagger/swagger'
#swagger generate server -f ../../../../swagger.yaml
swagger generate server -P models.Principal -f swagger.yaml