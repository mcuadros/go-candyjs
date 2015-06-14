time = CandyJS.require("time");
gin = CandyJS.require("github.com/gin-gonic/gin");

r = gin.Default()
r.GET("/back", CandyJS.proxy(function(c) {
    future = time.Date(2015, 10, 21, 4, 29 ,0, 0, time.UTC)
    now = time.Now()

    c.JSON(200, {
        "future": future.String(),
        "now": now.String(),
        "nsecs": future.sub(now),
    })
}))
r.Run(":8080") 