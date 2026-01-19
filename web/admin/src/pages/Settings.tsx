import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function Settings() {
    return (
        <div className="space-y-6">
            <h2 className="text-2xl font-bold">系统设置</h2>

            <Card>
                <CardHeader>
                    <CardTitle>关于</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div>
                        <h3 className="font-medium">AI Gateway</h3>
                        <p className="text-sm text-muted-foreground">
                            一个通用的 AI 网关服务，提供标准的 OpenAI 和 Anthropic 兼容接口，
                            实现不同 LLM 提供商之间的协议转换。
                        </p>
                    </div>
                    <div className="space-y-2 text-sm">
                        <p><strong>功能特性：</strong></p>
                        <ul className="list-disc list-inside text-muted-foreground space-y-1">
                            <li>双向协议兼容（OpenAI ↔ Anthropic）</li>
                            <li>流式响应支持 (Server-Sent Events)</li>
                            <li>工具/函数调用 (Tool Calling)</li>
                            <li>多模态支持（图片/视觉）</li>
                            <li>灵活的模型路由</li>
                            <li>负载均衡策略</li>
                        </ul>
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>环境变量配置</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                            <thead>
                                <tr className="border-b text-left text-muted-foreground">
                                    <th className="pb-2 font-medium">变量名</th>
                                    <th className="pb-2 font-medium">说明</th>
                                </tr>
                            </thead>
                            <tbody className="font-mono">
                                <tr className="border-b">
                                    <td className="py-2">DB_DSN</td>
                                    <td className="py-2 font-sans text-muted-foreground">MySQL 数据库连接字符串</td>
                                </tr>
                                <tr className="border-b">
                                    <td className="py-2">ADMIN_USER</td>
                                    <td className="py-2 font-sans text-muted-foreground">管理员用户名</td>
                                </tr>
                                <tr className="border-b">
                                    <td className="py-2">ADMIN_PASS</td>
                                    <td className="py-2 font-sans text-muted-foreground">管理员密码</td>
                                </tr>
                                <tr className="border-b">
                                    <td className="py-2">OPENAI_API_KEY</td>
                                    <td className="py-2 font-sans text-muted-foreground">OpenAI API 密钥</td>
                                </tr>
                                <tr>
                                    <td className="py-2">ANTHROPIC_API_KEY</td>
                                    <td className="py-2 font-sans text-muted-foreground">Anthropic API 密钥</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </CardContent>
            </Card>
        </div>
    )
}
