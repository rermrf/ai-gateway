import { useQuery } from '@tanstack/react-query'
import { modelApi } from '@/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function ModelPricingList() {
    const { data: models, isLoading } = useQuery({
        queryKey: ['models-pricing'],
        queryFn: modelApi.listWithPricing,
    })

    return (
        <Card>
            <CardHeader>
                <CardTitle>可用模型及价格</CardTitle>
            </CardHeader>
            <CardContent>
                {isLoading ? (
                    <div className="text-center py-8 text-muted-foreground">加载中...</div>
                ) : !models?.length ? (
                    <div className="text-center py-8 text-muted-foreground">暂无可用模型</div>
                ) : (
                    <div className="overflow-x-auto">
                        <table className="w-full">
                            <thead>
                                <tr className="border-b text-left text-sm text-muted-foreground">
                                    <th className="pb-3 font-medium">模型名称</th>
                                    <th className="pb-3 font-medium">输入价格 ($/1M tokens)</th>
                                    <th className="pb-3 font-medium">输出价格 ($/1M tokens)</th>
                                </tr>
                            </thead>
                            <tbody>
                                {models.map((model) => (
                                    <tr key={model.modelName} className="border-b last:border-0 hover:bg-muted/50 transition-colors">
                                        <td className="py-3 font-medium">{model.modelName}</td>
                                        <td className="py-3">
                                            {model.promptPrice > 0 ? (
                                                `$${model.promptPrice.toFixed(2)}`
                                            ) : (
                                                <span className="text-muted-foreground">-</span>
                                            )}
                                        </td>
                                        <td className="py-3">
                                            {model.completionPrice > 0 ? (
                                                `$${model.completionPrice.toFixed(2)}`
                                            ) : (
                                                <span className="text-muted-foreground">-</span>
                                            )}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </CardContent>
        </Card>
    )
}
