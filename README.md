# Remote js console

Remote js console is a simple microservice written in golang which can collect and display javascript errors or any other js events. You can log javascript errors or any other messages by adding a few lines of code to your html page. On server side a simple collector is running which saves all logs to mysql database. It also has a basic admin panel where you can see and filter latest logs.  


## Install

~~~~~
git clone 
go get
go build
~~~~~

## Setup

Edit config.json then start the server:

```
$ ./remote-js-console
2015/03/11 22:53:06 Starting server at localhost:8080 ...
```

You should see a message similar to the one above letting you know that logger is running on port 8080.

If server is running you can start collecting logs by adding few lines of code to your html:

~~~~~
<script>
//log message
function remoteLog(message) {
    var img = new Image();
    img.src = "http://localhost:8080/?msg=" + encodeURIComponent(message);
}

//log all errors
window.onerror = function (errorMsg, url, lineNumber, column, errorObj) {
    remoteLog('Error: ' + errorMsg + ' Script: ' + url + ' Line: ' + lineNumber + ' Column: ' + column + ' StackTrace: ' +  errorObj);
}
</script>
~~~~~

If everything works you can start the server to run in the background:
~~~~~
nohup ./remote-js-console > /dev/null & 
~~~~~

Or if you want to start on boot create /etc/init/remotejs.conf. Check upstart docs for more info.
