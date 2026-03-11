package define

const InfoCrawlTaskStatusDelete = 0
const InfoCrawlTaskStatusNormal = 1

const InfoCrawlRunStatusRunning = `running`
const InfoCrawlRunStatusSuccess = `success`
const InfoCrawlRunStatusFailed = `failed`

const InfoCrawlSseTypeStatus = `info_crawl_status`
const InfoCrawlSseTypeChunk = `info_crawl_chunk`
const InfoCrawlSseTypeDone = `info_crawl_done`
const InfoCrawlSseTypeError = `error`

const Crawl4AIStatusIdle = `idle`
const Crawl4AIStatusInstalling = `installing`
const Crawl4AIStatusReady = `ready`
const Crawl4AIStatusFailed = `failed`
