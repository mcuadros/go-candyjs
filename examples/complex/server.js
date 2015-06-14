time = CandyJS.require("time");
gin = CandyJS.require("github.com/gin-gonic/gin");

engine = gin.default()
engine.get("/back", CandyJS.proxy(function(ctx) {
    future = time.date(2015, 10, 21, 4, 29 ,0, 0, time.UTC)
    now = time.now()

    ctx.json(200, {
        "future": future.string(),
        "now": now.string(),
        "nsecs": future.sub(now),
    })
}))

engine.run(":8080") 