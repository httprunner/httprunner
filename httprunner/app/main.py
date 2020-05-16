from fastapi import FastAPI

from httprunner import __version__
from .routers import deps, debugtalk, debug

app = FastAPI()


@app.get("/hrun/version")
async def get_hrun_version():
    return {"code": 0, "message": "success", "result": {"HttpRunner": __version__}}


app.include_router(deps.router)
app.include_router(debugtalk.router)
app.include_router(debug.router)
