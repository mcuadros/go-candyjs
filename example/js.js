var circle = require('./circle.js');
console.log('The area of a circle of radius 4 is ' + circle.area(4));

var filesize = require("./filesize.min.js")
console.log(filesize(1234))


// Create proxy object.
var proxy = new Proxy({foo: 'bar'}, handler);

// Proxy object is then accessed normally.
print('foo' in proxy);
proxy.foo = "qux";
print(proxy.foo);
delete proxy.foo;
