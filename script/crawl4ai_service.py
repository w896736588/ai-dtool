#!/usr/bin/env python
# -*- coding: utf-8 -*-

"""
本地 Crawl4AI API 服务。
职责：
1. 提供健康检查接口
2. 接收多个 URL
3. 使用 Crawl4AI 抓取并返回 markdown
"""

import asyncio
import os
from typing import List

from fastapi import FastAPI
from pydantic import BaseModel
import uvicorn

from crawl4ai import AsyncWebCrawler


class CrawlRequest(BaseModel):
    urls: List[str]
    cache_mode: str = "bypass"
    word_count_threshold: int = 1


app = FastAPI(title="dtool-crawl4ai-service")


@app.get("/healthz")
async def healthz():
    """healthz 健康检查接口。"""
    return {"ok": True}


@app.post("/crawl")
async def crawl(request: CrawlRequest):
    """crawl 抓取多个网址并返回 markdown。"""
    results = []
    async with AsyncWebCrawler() as crawler:
        for url in request.urls:
            item = {"url": url, "success": False, "markdown": "", "title": "", "error": ""}
            try:
                result = await crawler.arun(url=url)
                item["success"] = bool(getattr(result, "success", True))
                item["markdown"] = getattr(result, "markdown", "") or ""
                item["title"] = getattr(result, "title", "") or ""
                if not item["success"]:
                    item["error"] = getattr(result, "error_message", "") or "抓取失败"
            except Exception as exc:
                item["error"] = str(exc)
            results.append(item)
    return {"data": results}


def main():
    """main 读取环境变量并启动 uvicorn。"""
    host = os.getenv("CRAWL4AI_HOST", "127.0.0.1")
    port = int(os.getenv("CRAWL4AI_PORT", "11235"))
    asyncio.set_event_loop_policy(asyncio.WindowsProactorEventLoopPolicy())
    uvicorn.run(app, host=host, port=port, log_level="warning")


if __name__ == "__main__":
    main()
