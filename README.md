# events

Start Server
`
go run .
`
Request Sample 
Post Event
URL -> http://localhost:1323
`
{
    "type":"create",
    "data":{"key":"key3","value":"value1"}
}

`

Get Event
URL -> http://localhost:1323


Get History
URL -> http://localhost:1323/history/key3


Run Test
`
go test .
`