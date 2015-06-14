http = CandyJS.require("net/http");
ioutil = CandyJS.require("io/ioutil");

resp = http.get("http://localhost:8080/back");

if (resp.statusCode == 200) {
    json = ioutil.readAll(resp.body);
    obj = JSON.parse(json);

    print("Back to the future date:", obj.future);
    print("Current date:", obj.future);
    print("Back to the Future day is on: " + obj.nsecs + " nsecs!");
} else {
    print("Request failed, status code:", resp.statusCode);
}