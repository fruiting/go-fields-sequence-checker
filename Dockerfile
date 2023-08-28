FROM golang:1.19

RUN go install github.com/fruiting/go-fields-sequence-checker@latest

CMD ["go-fields-sequence-checker"]
