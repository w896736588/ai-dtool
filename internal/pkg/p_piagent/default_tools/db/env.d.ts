// 编辑器类型声明（仅供 VS Code / tsc 静态检查，不影响 Pi 运行时）。
declare module '@earendil-works/pi-coding-agent' {
  export interface ExtensionAPI {
    registerTool(config: {
      name: string;
      description: string;
      parameters: any;
      execute: (
        toolCallId: string,
        params: any,
        signal?: any,
        onUpdate?: any,
        ctx?: any,
      ) => Promise<{ content: Array<{ type: string; text: string }>; details: any }>;
    }): void;
  }
}
