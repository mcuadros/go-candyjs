http = CandyJS.require("net/http");
ioutil = CandyJS.require("io/ioutil");

resp = http.Get("http://localhost:8080/back");

if (resp.StatusCode == 200) {
    json = ioutil.ReadAll(resp.Body);
    obj = JSON.parse(json);

    print("Back to the future date:", obj.future);
    print("Current date:", obj.future);
    print("Back to the Future day is on: " + obj.nsecs + " nsecs!");
} else {
    print("Request failed, status code:", resp.StatusCode);
}