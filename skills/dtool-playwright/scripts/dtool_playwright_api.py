import json
import os
import sys
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../dtool-common/scripts'))

from api_common import BASE_URL, TOKEN, call_api


def browser_profile_open(base_url: str, token: str, smart_link_id: int,
                         account: str = "", open_type: int = 0,
                         reuse_if_open: bool = True, enable_mcp: bool = False):
    # 同步全局配置，使 call_api 使用正确的地址和令牌
    import api_common
    api_common.BASE_URL = base_url
    api_common.TOKEN = token
    payload = {
        "smart_link_id": smart_link_id,
        "account": account,
        "open_type": open_type,
        "reuse_if_open": reuse_if_open,
        "enable_mcp": enable_mcp,
    }
    return call_api("/api/ai/browser/session/open", payload)


def extract_user_data_dir(open_result: dict) -> str:
    return open_result.get("data", {}).get("user_data_dir", "")


def extract_executable_path(open_result: dict) -> str:
    return open_result.get("data", {}).get("native_playwright", {}).get("executable_path", "")


def extract_mcp_info(open_result: dict) -> dict:
    """提取 MCP 连接信息，返回 mcp 字段内容或空字典。"""
    return open_result.get("data", {}).get("mcp", {})


def is_mcp_mode(open_result: dict) -> bool:
    """判断响应是否为 MCP 模式。"""
    return extract_mcp_info(open_result).get("enabled", False)


# ── DOM 分析辅助函数（旧模式，供 AI 在 Playwright 会话中使用） ──

JS_EXTRACT_INTERACTIVE_ELEMENTS = """() => {
    const result = [];
    document.querySelectorAll(
        'button, a, input, select, textarea, [role="button"], [role="link"], [role="tab"], [role="menuitem"]'
    ).forEach(el => {
        result.push({
            tag: el.tagName.toLowerCase(),
            text: (el.textContent || '').trim().substring(0, 100),
            type: el.type || '',
            placeholder: el.placeholder || '',
            href: el.href || '',
            role: el.getAttribute('role') || '',
            aria_label: el.getAttribute('aria-label') || '',
            name: el.name || '',
            id: el.id || '',
            disabled: el.disabled || false,
            visible: el.offsetParent !== null,
        });
    });
    return result;
}"""

JS_EXTRACT_PAGE_TEXT = """(max_length = 5000) => {
    const text = document.body.innerText.substring(0, max_length);
    return text;
}"""


def build_python_playwright_snippet(user_data_dir: str, goto_url: str = "",
                                     executable_path: str = "") -> str:
    goto_line = ""
    if goto_url:
        goto_line = f'    page.goto(r"{goto_url}")\n'
    exec_path_line = ""
    if executable_path:
        exec_path_line = f'        executable_path=r"{executable_path}",\n'
    return (
        "from playwright.sync_api import sync_playwright\n\n"
        f'user_data_dir = r"{user_data_dir}"\n'
        f'executable_path = r"{executable_path}"\n\n'
        "with sync_playwright() as p:\n"
        "    context = p.chromium.launch_persistent_context(\n"
        "        user_data_dir=user_data_dir,\n"
        f"{exec_path_line}"
        "        headless=False,\n"
        "        no_viewport=True,\n"
        "    )\n"
        "    page = context.pages[0] if context.pages else context.new_page()\n"
        f"{goto_line}"
        "    # 用无障碍树分析页面（毫秒级，无需截图）\n"
        "    snapshot = page.accessibility.snapshot()\n"
        "    print(snapshot)\n"
        "    # 用 JS 提取可交互元素\n"
        f"    elements = page.evaluate({repr(JS_EXTRACT_INTERACTIVE_ELEMENTS)})\n"
        "    print(elements)\n"
    )


if __name__ == "__main__":
    base_url = "http://127.0.0.1:17170"
    token = "请替换为真实Token"

    try:
        # MCP 模式示例
        open_result = browser_profile_open(
            base_url=base_url,
            token=token,
            smart_link_id=12,
            account="tester",
            enable_mcp=True,
        )
        print("open_result:")
        print(json.dumps(open_result, ensure_ascii=False, indent=2))

        if is_mcp_mode(open_result):
            mcp_info = extract_mcp_info(open_result)
            print(f"\nMCP 模式已启用")
            print(f"  SSE 端点: {mcp_info.get('sse_endpoint')}")
            print(f"  Message 端点: {mcp_info.get('msg_endpoint')}")
        else:
            user_data_dir = extract_user_data_dir(open_result)
            if user_data_dir:
                executable_path = extract_executable_path(open_result)
                goto_url = open_result.get("data", {}).get("site", {}).get("url", "")
                print("\npython_playwright_snippet:")
                print(build_python_playwright_snippet(user_data_dir, goto_url, executable_path))

    except Exception as exc:
        print(f"请求失败: {exc}")
