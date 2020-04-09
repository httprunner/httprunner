from typing import Dict, Text

from pydantic import BaseModel, Field

from httprunner.schema import common


class Api(BaseModel):
    name: common.Name
    request: common.Request
    variables: common.Variables = {}
    base_url: common.BaseUrl = ""
    setup_hooks: common.Hook = []
    teardown_hooks: common.Hook = []
    extract: Dict[Text, Text] = {}
    validation: common.Validate = Field([], alias="validate")
