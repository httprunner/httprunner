import logging
import time
import yaml
from typing import List
import json
import requests
import os
import re
from collections import Counter


import sys


def ret():
    # 获取脚本名称
    script_name = sys.argv[0]

    # 获取其他命令行参数
    args = sys.argv[1:]
    print("脚本名称:", script_name)
    print("命令行参数:", args)
    return "sh脚本执行参数"+args


config = {}


def init_config():
    global config
    with open('mnt/auto-config/config.conf') as f:
        config = json.load(f)


def get_config_value(key):
    cls = os.getenv("CLUSTER")
    if not config:
        init_config()
    if key == "CLUSTER":
        return os.getenv(key)

    if key == "MESHID":
        tmpkey = '{}-{}'.format(cls, "meshID").upper()
        meshID = os.getenv(tmpkey)
        if meshID == None or meshID == "":
            meshID = get_meshID(cls)
            os.environ[tmpkey] = meshID
        return meshID

    if key == "EXTERNAL_IP":
        return get_cluster_ip_from_configmap(cls)

    if key == "APP_IMAGE":
        return fetch_chart_image("asm-application", "")

    if key == "INGRESS_IMAGE":
        return fetch_chart_image("asm-ingress", "")

    if key == "HOST":
        host = os.getenv("CCE_API_DOMAIN")
        return f"https://{host}"

    if key == "NAMESPACE":
        val = os.getenv("NAMESPACE")
        return val

    if key == "ORGANIZATION":
        val = os.getenv("ORGANIZATION")
        return val

    if key == "NCCP_CLUSTER":
        host = os.getenv("NCCP_CLUSTER")
        return host

    return config.get(key)


def ENV_VAL(key):
    return get_config_value(key)


def get_mnt_cookie():
    f = open('/mnt/cookie', 'r')
    content = f.read()
    f.close()
    return content


def get_chartid(body, name, version) -> str:
    data_str = json.dumps(body)
    json_list = json.loads(data_str)
    for key in json_list:
        if key["name"] == name and key["version"] == version:
            return key["id"]
    return ""


def get_cookie():
    host = get_host_info()
    uri = "{}/backend/cce/tm/apis/cce/v1/users/login".format(host)
    headers = {
        "Content-Type": "application/json;charset=utf-8"
    }
    data = json.dumps({
        "name": ENV_VAL("USERNAME"),
        "password": ENV_VAL("PASSWORD")
    })
    try:
        response = requests.post(uri, data=data, headers=headers, verify=False)
        if response.status_code == 200:
            cookie = response.headers["Set-Cookie"]
            return cookie
    except Exception as err:
        print(f"Error: {err}")
        return err
    return ""


def get_meshID(cls):
    host = get_host_info()
    uri = "{}/backend/asm/apiserver/v2/projects/default/meshes".format(host)
    cookie = get_cookie()
    if cookie == "":
        return "cookie not available"

    headers = {
        "Content-Type": "application/json;charset=utf-8",
        "Cookie": cookie
    }

    try:
        response = requests.get(uri, headers=headers, verify=False)
        if response.status_code == 200:
            resp = response.json()
            meshes = resp.get("meshes", [])
            for mesh in meshes:
                if mesh["meshName"] == cls:
                    return mesh["meshId"]
        else:
            print(
                f"Error - code: {response.status_code}, body: {response.text}")
            return None
    except Exception as err:
        print(f"Error: {err}")
    return ""


def get_config_map(cluster, namespace, name):
    host = get_host_info()
    id_token = get_token()
    uri = f"{host}/api/v1/namespaces/{namespace}/configmaps/{name}"
    headers = {
        "Content-Type": "application/json;charset=utf-8",
        "X-Cluster-ID": cluster,
        "Authorization": "bearer " + id_token,
    }

    try:
        response = requests.get(uri, headers=headers, verify=False)
        if response.status_code == 200:
            gc_resp = response.json()
            return gc_resp.get("data", {})
        else:
            resp = response.json()
            raise Exception(
                f"Error - {resp.get('message', '')} {resp.get('reason', '')}")
    except Exception as err:
        return err


def get_cluster_ip_from_configmap(cluster):
    time.sleep(3)
    try:
        configmap = get_config_map("primary", "nccp-alert", "nccp-tool-config")
        if "asm-gateway_ip.json" not in configmap:
            return Exception("获取集群IP失败")

        configs_map = json.loads(configmap["asm-gateway_ip.json"])
        print(configs_map)
        if cluster not in configs_map:
            return Exception(f"获取{cluster}集群IP失败")

        return configs_map[cluster][0]

    except Exception as err:
        return err

# def opt_yml(file_name, namespace):
#     # 读取yml文件
#     with open(file_name, "r", encoding="utf-8") as f:
#         cfg = yaml.load(f, Loader=yaml.FullLoader)
#     if namespace != "null":
#         cfg["global"]["namespace"] = namespace
#     return cfg


def http_compatible_application_values(appname, ns, image, port, service_port, chartvalues):
    chart_values = json.loads(chartvalues)

    env_data = [
        {"name": "NCCP_CLUSTER", "value": "primary"},
        {"name": "ns", "value": ns},
        {"name": "appname", "value": appname}
    ]

    chart_values["appname"] = appname
    chart_values["containers"][0]["image"] = image
    chart_values["containers"][0]["ports"][0]["containerPort"] = service_port
    chart_values["containers"][0]["livenessProbe"] = None
    chart_values["containers"][0]["readinessProbe"] = None
    chart_values["service"][0]["port"] = port
    chart_values["service"][0]["targetPort"] = service_port
    # 读取yml文件
    # cfg = template_data.application_values(appname,image,port,service_port)
    if appname != "null":
        chart_values["appname"] = appname
        for key in chart_values["containers"]:
            key["env"] = env_data

    return chart_values


def upd_tcp_app_yml(appname, image, port, service_port, server_addr, is_long_chain, send_heartbeat_gap, chartvalues):
    env_data = [
        {"name": "NCCP_CLUSTER", "value": "primary"},
        {"name": "LISTEN_PORT", "value": "8888"},
        {"name": "SERVER_ADDR", "value": server_addr},
        {"name": "SERVER_NAME", "value": appname},
        {"name": "IS_LONG_CHAIN", "value": is_long_chain},
        {"name": "SEND_HEARTBEAT_GAP", "value": send_heartbeat_gap},
        {"name": "appname", "value": appname}
    ]

    chart_values = json.loads(chartvalues)
    chart_values["appname"] = appname
    chart_values["containers"][0]["image"] = image
    chart_values["containers"][0]["ports"][0]["containerPort"] = service_port
    chart_values["containers"][0]["livenessProbe"] = None
    chart_values["containers"][0]["readinessProbe"] = None
    chart_values["service"][0]["port"] = port
    chart_values["service"][0]["targetPort"] = service_port
    # cfg = template_data.application_values(appname,image,port,service_port)

    # 读取yml文件
    if appname != "null":
        chart_values["appname"] = appname
        for key in chart_values["containers"]:
            key["env"] = env_data
    return chart_values

# def upd_tcp_gateway_yml(file_name, port, ns):

#     # 读取yml文件
#     with open(file_name, "r", encoding="utf-8") as f:
#         cfg = yaml.load(f, Loader=yaml.FullLoader)
#         cfg["gateway"]["externalPort"] = port
#         cfg["gateway"]["name"] = "tcp-{}".format(port)
#         cfg["global"]["namespace"] = ns

#     return cfg

# def upd_tcp_route_yml(file_name, port, ns, servicename):

#     # 读取yml文件
#     with open(file_name, "r", encoding="utf-8") as f:
#         cfg = yaml.load(f, Loader=yaml.FullLoader)
#         cfg["gateway"]["externalPort"] = port
#         cfg["gateway"]["name"] = "tcp-{}".format(port)
#         cfg["global"]["namespace"] = ns
#         cfg["global"]["serviceName"] = servicename

#     return cfg

# def opt_ingress_yml(file_name, ver):
#      # 读取yml文件
#     with open(file_name, "r", encoding="utf-8") as f:
#         cfg = yaml.load(f, Loader=yaml.FullLoader)
#     if ver == "old":
#         cfg["image"] = "ncpdev.gf.com.cn/cce/proxyv2:1.8.6-r2-20230119163403"

#     return cfg

# def get_access_url(file_name):
#     # 读取yml文件
#     with open(file_name, "r", encoding="utf-8") as f:
#         cfg = yaml.load(f, Loader=yaml.FullLoader)

#     host = ENV_VAL("EXTERNAL_IP")
#     port = cfg["gateway"]["externalPort"]
#     path = ""
#     # for key in cfg["gateway"]["externalIPs"]:
#     #     host = key
#     #     break
#     for key in cfg["global"]["httpservicelist"]:
#         path = key["path"]
#         break

#     return "http://{}:{}{}".format(host,port,path)


def get_host_info() -> str:
    return ENV_VAL("HOST")


def get_token() -> str:
    host = get_host_info()
    url = host + "/auth/realms/CCE/protocol/openid-connect/token"
    headers = {"content-type": "application/x-www-form-urlencoded"}
    param = {
        "client_id": "cce-client",
        "response_type": "code",
        "grant_type": "password",
        "scope": "openid",
        "username": ENV_VAL("USERNAME"),
        "password": ENV_VAL("PASSWORD"),
    }
    response = requests.post(url, data=param, headers=headers, verify=False)
    if response.status_code == 200:
        return response.json()["id_token"]


def create_namespace(cluster, organization, namespace) -> str:
    host = get_host_info()
    id_token = get_token()
    url = host + "/apis/cce/v1/organizations/" + organization + "/namespaces"
    headers = {
        "Authorization": "bearer " + id_token,
        "X-Cluster-ID": cluster,
        "Content-Type": "application/json;charset=utf-8",
    }
    payload = {"name": namespace}
    response = requests.post(url, json=payload, headers=headers, verify=False)
    return response.json()


def del_configmap(cluster, namespace):
    host = get_host_info()
    id_token = get_token()
    url = host + "/api/v1/namespaces/" + namespace + "/configmaps"
    headers = {
        "Authorization": "bearer " + id_token,
        "X-Cluster-ID": cluster,
        "Content-Type": "application/json;charset=utf-8",
    }
    response = requests.delete(url, headers=headers, verify=False)
    return response.json()


def del_secrets(cluster, namespace):
    host = get_host_info()
    id_token = get_token()
    url = host + "/api/v1/namespaces/" + namespace + "/secrets"
    headers = {
        "Authorization": "bearer " + id_token,
        "X-Cluster-ID": cluster,
        "Content-Type": "application/json;charset=utf-8",
    }
    response = requests.delete(url, headers=headers, verify=False)
    return response.json()


def del_namespace(cluster, organization, namespace):
    host = get_host_info()
    id_token = get_token()
    url = (
        host + "/apis/cce/v1/organizations/" + organization + "/namespaces/" + namespace
    )
    headers = {
        "Authorization": "bearer " + id_token,
        "X-Cluster-ID": cluster,
        "Content-Type": "application/json;charset=utf-8",
    }
    response = requests.delete(url, headers=headers, verify=False)
    return response.json()


def get_charts(organization):
    host = get_host_info()
    id_token = get_token()
    url = host + "/v2/charts"
    headers = {
        "Authorization": "bearer " + id_token,
        "Content-Type": "application/json;charset=utf-8",
        "Organization": organization,
    }
    payload = {"public": False}
    response = requests.get(url, payload=payload,
                            headers=headers, verify=False)
    return response.json()


def get_releases(cluster, organization, namespace):
    host = get_host_info()
    id_token = get_token()
    url = host + "/v2/releases"
    headers = {
        "Authorization": "bearer " + id_token,
        "Content-Type": "application/json;charset=utf-8",
        "Organization": organization,
    }
    params = {"namespace": namespace, "cluster": cluster}
    response = requests.get(url, params=params, headers=headers, verify=False)
    res = {"id_token": id_token, "response": response.json()}
    data_str = json.dumps(res)
    return data_str


def del_releases(cluster, organization, namespace):
    res = get_releases(cluster, organization, namespace)
    data = json.loads(res)
    host = get_host_info()
    url = host + "/v2/releases"
    headers = {
        "Authorization": "bearer " + data["id_token"],
        "Content-Type": "application/json;charset=utf-8",
        "Organization": organization,
    }
    params = {"namespace": namespace, "cluster": cluster}

    for v in data["response"]:
        params["name"] = v["name"]
        response = requests.delete(
            url, params=params, headers=headers, verify=False)
    return "True"


def del_deployment(cluster, namespace, name):
    host = get_host_info()
    id_token = get_token()

    url = host + "/apis/apps/v1/namespaces/" + namespace + "/deployments/" + name
    headers = {
        "Authorization": "bearer " + id_token,
        "X-Cluster-ID": cluster,
        "Content-Type": "application/json;charset=utf-8",
    }
    response = requests.delete(url, headers=headers, verify=False)
    return response.json()


def ns_sidecar_enable(cluster, namespace):
    host = get_host_info()
    id_token = get_token()

    url = host + "/api/v1/namespaces/" + namespace
    headers = {
        "Authorization": "bearer " + id_token,
        "X-Cluster-ID": cluster,
        "Content-Type": "application/json-patch+json",
    }
    params = [
        {"op": "add", "path": "/metadata/labels/istio-injection", "value": "enabled"}
    ]
    response = requests.patch(url, headers=headers, json=params, verify=False)
    return response.json()


def delete_release(baseurl, id_token, organization, cluster, namespace, name):
    url = "{}/v2/releases?name={}&cluster={}&namespace={}".format(
        baseurl, name, cluster, namespace
    )
    headers = {
        "Content-Type": "application/json;charset=utf-8",
        "Authorization": "bearer " + id_token,
        "Organization": organization,
    }
    response = requests.delete(url, headers=headers, verify=False)
    if response.status_code == 200:
        time.sleep(5)  # 等待删除实例
    return response.status_code


def delete_deployment(baseurl, id_token, organization, cluster, namespace, name):
    url = "{}/apis/apps/v1/namespaces/{}/deployments/{}".format(
        baseurl, namespace, name
    )
    headers = {
        "Content-Type": "application/json;charset=utf-8",
        "Authorization": "bearer " + id_token,
        "Organization": organization,
        "X-Cluster-ID": cluster,
    }
    response = requests.delete(url, headers=headers, verify=False)
    return response.status_code


def wait_delete_namespace(baseurl, id_token, organization, cluster, namespace):
    url = "{}/apis/cce/v1/organizations/{}/namespaces".format(
        baseurl, organization
    )
    headers = {
        "Content-Type": "application/json;charset=utf-8",
        "Authorization": "bearer " + id_token,
        "X-Cluster-ID": cluster,
    }
    begin_time = time.time()
    while time.time() - begin_time < 300:
        try:
            response = requests.get(url, headers=headers, verify=False)
            print(response.status_code)
            if response.status_code == 200:
                # 提取并解析命名空间信息
                data = response.json()
                namespaces = data['items']
                flag = True

                for namespace_data in namespaces:
                    namespace_name = namespace_data['namespace']
                    if namespace_name == namespace:
                        flag = False
                        break

                if flag == True:
                    return "True"

            time.sleep(5)
        except requests.RequestException as e:
            print(f"Request failed: {e}")
            time.sleep(5)
    return "False"


def retry_access(url):
    begin_time = time.time()
    while time.time() - begin_time < 300:
        try:
            response = requests.get(url)
            if response.status_code == 200:
                return "True"  # Successful response
            time.sleep(5)
            begin_time = time.time()
        except requests.RequestException as e:
            print(f"Request failed: {e}")
    return "False"  # All retries failed


def waiting_deployment(baseurl, id_token, organization, cluster, namespace, name, begin_time):
    url = "{}/apis/apps/v1/namespaces/{}/deployments/{}".format(
        baseurl, namespace, name
    )

    headers = {
        "Content-Type": "application/json;charset=utf-8",
        "Authorization": "bearer " + id_token,
        "Organization": organization,
        "X-Cluster-ID": cluster,
    }
    num = 0
    begin_time = time.time()
    while time.time() - begin_time < 600:
        try:
            response = requests.get(url, headers=headers, verify=False)
            if response.status_code != 200:
                num = num + 1
                if num > 3:
                    return f"{response.status_code}"
                time.sleep(5)
                continue

            resp = response.json()
            replicas = resp["status"].get("replicas", 0)
            ready_replicas = resp["status"].get("readyReplicas", 0)

            if replicas and ready_replicas:
                return "True"
            time.sleep(5)

        except requests.RequestException as e:
            print(f"Request failed: {e}")

    return "False"  # All retries failed


def try_http_service(baseurl, begin_time):
    url = baseurl

    headers = {
        "Content-Type": "application/json;charset=utf-8",
    }

    begin_time = time.time()
    while time.time() - begin_time < 300:
        try:
            response = requests.get(url, headers=headers, verify=False)
            if response.status_code == 200:
                return "True"
            time.sleep(5)
        except requests.RequestException as e:
            print(f"Request failed: {e}")

    return "False"  # All retries failed


def now():
    return int(time.time())


def init_secret_test_data(cluster, namespace):
    time.sleep(6)
    host = get_host_info()
    id_token = get_token()

    url = host + "/nccp/secret/api/create"
    headers = {
        "Authorization": "bearer " + id_token,
        "X-Cluster-ID": cluster,
        "Content-Type": "application/json",
    }

    result = "初始化secret数据成功"
    i = 1
    while i <= 5:
        params = {
            "cluster": cluster,
            "namespace": namespace,
            "name": "secret{}".format(i),
            "data": {"key{}".format(i): "value{}".format(i)},
        }

        response = requests.post(url, headers=headers,
                                 json=params, verify=False)
        if response.json()["code"] != 0:
            result = "初始化secret数据失败, err: " + response.json()["data"]
            break
        i = i + 1

    return result


def test_630_tcp_gateway():
    ip = ENV_VAL("EXTERNAL_IP")
    return test_tcp_con.test_630_tcp_gateway(ip)


def http_route_values(external_ip, external_port, host, name, path, port, svcname, namespace, chartvalues):
    chart_values = json.loads(chartvalues)

    chart_values["gateway"]["externalIPs"] = [external_ip]
    chart_values["gateway"]["externalPort"] = external_port
    chart_values["gateway"]["host"] = host
    chart_values["gateway"]["protocol"] = 'http'
    chart_values["gateway"]["name"] = name
    chart_values["global"]["httpservicelist"] = [{
        "path": path,
        "port": port,
        "svcname": svcname
    }]
    chart_values["global"]["namespace"] = namespace

    return chart_values


def http_serviceentry_values(addresses_ip, appname, endpoints_ip, endpoints_port, host, chartvalues):
    chart_values = json.loads(chartvalues)

    chart_values["addresses"] = [addresses_ip]
    chart_values["appname"] = appname
    chart_values["endpoints"] = [{
        "ip": endpoints_ip,
        "port": endpoints_port,
    }]
    chart_values["host"] = host

    return chart_values


def application_values(appname, image, port, service_port, chartvalues):
    chart_values = json.loads(chartvalues)

    chart_values["appname"] = appname
    chart_values["containers"][0]["image"] = image
    chart_values["containers"][0]["ports"][0]["containerPort"] = service_port
    chart_values["service"][0]["port"] = port
    chart_values["service"][0]["targetPort"] = service_port

    return chart_values


def http_gateway_values(external_ip, external_port, host, name, namespace, chartvalues):
    chart_values = json.loads(chartvalues)

    chart_values["gateway"]["externalIPs"] = [external_ip]
    chart_values["gateway"]["externalPort"] = external_port
    chart_values["gateway"]["protocol"] = "http"
    chart_values["gateway"]["host"] = host
    chart_values["gateway"]["name"] = name
    chart_values["global"]["namespace"] = namespace

    return chart_values


def ingress_values(as_cluster_id, mesh_id, cluster_id, val, chartvalues):
    chart_values = json.loads(chartvalues)
    chart_values["env"]["ISTIO_META_ASM_CLUSTER_ID"] = as_cluster_id
    chart_values["env"]["ISTIO_META_ASM_MESH_ID"] = mesh_id
    chart_values["env"]["ISTIO_META_CLUSTER_ID"] = cluster_id
    if val == "old":
        chart_values["image"] = ENV_VAL("INGRESS_IMAGE_OLD")
    else:  # new
        chart_values["image"] = ENV_VAL("INGRESS_IMAGE")
    return chart_values


def tcp_route_values(external_ip, external_port, host, namespace, service_name, service_port, chartvalues):
    name = "tcp-{}".format(external_port)
    chart_values = json.loads(chartvalues)
    chart_values["gateway"]["externalIPs"] = [external_ip]
    chart_values["gateway"]["externalPort"] = external_port
    chart_values["gateway"]["protocol"] = 'tcp'
    chart_values["gateway"]["host"] = host
    chart_values["gateway"]["name"] = name
    chart_values["global"]["namespace"] = namespace
    chart_values["global"]["serviceName"] = service_name
    chart_values["global"]["servicePort"] = service_port

    return chart_values


def tcp_gateway_values(external_ip, external_port, host, namespace, chartvalues):
    name = "tcp-{}".format(external_port)
    chart_values = json.loads(chartvalues)

    # str = response_data_json["charts"][0]["values"]
    chart_values = response_data["charts"][0]["values"]
    chart_values["gateway"]["externalIPs"] = [external_ip]
    chart_values["gateway"]["externalPort"] = external_port
    chart_values["gateway"]["protocol"] = "tcp"
    chart_values["gateway"]["host"] = host
    chart_values["gateway"]["name"] = name
    chart_values["global"]["namespace"] = namespace
    return chart_values


def json_to_yaml(values):
    data = yaml.safe_load(values)
    return data


def fetch_chart_info(chart_name, organization):
    host = get_host_info()
    id_token = get_token()
    # 构建请求头
    headers = {
        "Content-Type": "application/json;charset=utf-8",
        "Authorization": f"bearer {id_token}",
        "Organization": organization
    }

    # 构建请求参数
    params = {
        "public": True,
        "name": chart_name
    }
    response = {}
    # 发起 GET 请求
    begin_time = time.time()
    while time.time() - begin_time < 100:
        try:
            response = requests.get(
                f"{host}/v2/charts", headers=headers, params=params)
            if response.status_code == 200:
                break
            time.sleep(5)
        except requests.RequestException as e:
            print(f"Request failed: {e}")
            time.sleep(5)
    # 解析响应
    return response


def fetch_chart_image(chart_name, organization):
    response_data = fetch_chart_info(chart_name, organization)
    response_data = response_data.json()
    if len(response_data.get("charts", [])) > 0:
        chart_values_str = response_data["charts"][0]["values"]
        chart_values = json.loads(chart_values_str)

        if chart_name == "asm-ingress" and chart_values["image"] != "":
            return chart_values["image"]
        if chart_name == "asm-application" and chart_values["containers"][0]["image"] != None:
            return chart_values["containers"][0]["image"]
    else:
        print("Validation Failed")
    return response_data


def send_request_with_retry(url, method, data=None, headers=None, max_retries=3):
    for retry in range(max_retries):
        try:
            response = requests.request(
                method, url, data=data, headers=headers)
            if response.status_code == 200:
                return response
        except Exception as e:
            print(f"Request failed on attempt {retry + 1}: {str(e)}")
        time.sleep(2)  # 添加重试之间的延迟
    raise Exception(
        f"Max retries ({max_retries}) reached. Unable to send request.")


# 独立验证方法执行
if __name__ == "__main__":
    m = http_gateway_values("10.128.12.40", 30836, "",
                            "", "nccp-auto-test-asm")
    print(m)
