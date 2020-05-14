import contextlib
import logging
import sys
from io import StringIO

from fastapi import APIRouter
from starlette.requests import Request

router = APIRouter()


@contextlib.contextmanager
def stdout_io(stdout=None):
    old = sys.stdout
    if stdout is None:
        stdout = StringIO()
    sys.stdout = stdout
    yield stdout
    sys.stdout = old


@router.post("/hrun/debug/debugtalk_py", tags=["debugtalk"])
async def debug_python(request: Request):
    body = await request.body()

    if request.headers.get("content-transfer-encoding") == "base64":
        # TODO: decode base64
        pass

    resp = {"code": 0, "message": "success", "result": ""}
    try:
        with stdout_io() as s:
            exec(body, globals())
            output = s.getvalue()
            resp["result"] = output
    except Exception as ex:
        resp["code"] = 1
        resp["message"] = "fail"
        resp["result"] = str(ex)
        logging.error(resp)

    return resp
