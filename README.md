# relay

relay tcp stream.

usage:

open browser, set proxy to localhost:1080
open relay1: ./main -faddr ":1080" -baddr ":1081" -brc "123"
open relay2: ./main -faddr ":1081" -baddr ":1082" -frc "123" -brc "123"
open relay3: ./main -faddr ":1082" -baddr ":1083" -frc "123"
open sock5: ./sock5 -addr ":1083"

browser will send data to relay1,
relay1 will send data to relay2,
relay2 will send data to sock5.
sock5 : go get github.com/MDGSF/sock5

you can use arbitrary numbers of relay.