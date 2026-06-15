"""
更新工作流程任务步骤状态。

用法:
    from update_workflow_status import update_workflow_status
    update_workflow_status(base_url='http://localhost:17170', token='temptoken', workflow_id=69, step='requirement', status='running')

支持的 step 值：
    - 固定步骤: requirement-fetch, task-config, issue_fix
    - 开发步骤: requirement, design, api-dev, api-test-fix, code-review, browser-test
    - 文档步骤: plain_text_requirement, design_plan_requirement
    - 自定义步骤: custom_{id} (如 custom_10)
"""

import json
import urllib.request
import urllib.error

VALID_STATUSES = {"pending", "running", "completed"}


def _post(base_url, token, path, data):
    url = base_url.rstrip("/") + path
    req = urllib.request.Request(
        url,
        data=json.dumps(data, ensure_ascii=False).encode("utf-8"),
        headers={"Content-Type": "application/json; charset=utf-8", "Token": token},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        raise RuntimeError(f"HTTP {e.code}: {e.read().decode('utf-8', errors='replace')}")
    except urllib.error.URLError as e:
        raise RuntimeError(f"请求失败: {e.reason}")


def update_workflow_status(base_url, token, workflow_id, step, status):
    """更新工作流程指定步骤的状态。

    Args:
        base_url: dtool 服务地址
        token: 认证令牌
        workflow_id: 工作流程 ID
        step: 步骤 key（固定步骤/自定义步骤 custom_{id} 均支持）
        status: 状态值（pending/running/completed）
    """
    if status not in VALID_STATUSES:
        raise ValueError(f"无效状态 '{status}'，合法值: {', '.join(sorted(VALID_STATUSES))}")

    resp = _post(base_url, token, "/api/task/workflow/node-status/update", {
        "workflow_id": workflow_id,
        "step": step,
        "status": status,
    })
    if resp.get("ErrCode") != 0:
        raise RuntimeError(resp.get("ErrMsg", "更新节点状态失败"))
