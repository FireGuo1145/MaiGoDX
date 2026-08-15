import { useEffect, useRef, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import api from '@/lib/api'
import type { MetadataItem } from '@/types'
import { apiErrorMessage } from '@/types'

const kinds = [{ id: 'music', label: '歌曲列表' }, { id: 'partner', label: '搭档列表' }, { id: 'ticket', label: '功能票列表' }, { id: 'chara', label: '旅行伙伴列表' }]
export function MetadataPanel({ onChanged }: { onChanged: () => Promise<void> }) {
  const [kind, setKind] = useState('music')
  const [data, setData] = useState<Record<string, MetadataItem[]>>({})
  const [status, setStatus] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const fileRef = useRef<HTMLInputElement>(null)
  const items = data[kind] || []
  const refresh = async () => { try { const result = await api.getMetadata(); if (result.success) setData(result.metadata || {}) } catch (error) { setStatus(apiErrorMessage(error)) } }
  useEffect(() => { void refresh() }, [])
  const update = (index: number, field: 'id' | 'name', value: string) => { const next = [...items]; next[index] = { ...next[index], [field]: field === 'id' ? Number(value) : value }; setData({ ...data, [kind]: next }) }
  const add = () => setData({ ...data, [kind]: [...items, { id: 0, name: '' }] })
  const remove = (index: number) => setData({ ...data, [kind]: items.filter((_, i) => i !== index) })
  const save = async () => { setBusy(true); setStatus(null); try { const result = await api.saveMetadata(kind, items); if (!result.success) throw new Error(result.message || '保存失败'); await refresh(); await onChanged(); setStatus(result.message || '保存成功') } catch (error) { setStatus(apiErrorMessage(error)) } finally { setBusy(false) } }
  const importXML = async (file?: File) => { if (!file) return; setBusy(true); setStatus(null); try { const result = await api.importMetadata(file); if (!result.success) throw new Error(result.message || '导入失败'); await refresh(); await onChanged(); setStatus(result.message || '导入成功') } catch (error) { setStatus(apiErrorMessage(error)) } finally { setBusy(false); if (fileRef.current) fileRef.current.value = '' } }
  return (
    <Card className="border-slate-800 bg-slate-900">
      <CardHeader><CardTitle className="text-white">站点元数据</CardTitle><p className="text-sm text-slate-400">管理 ID 与名称的显示映射。仅用于门户展示，不会参与存档读写。</p></CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-wrap gap-2">{kinds.map((entry) => <Button key={entry.id} type="button" size="sm" variant={kind === entry.id ? 'default' : 'outline'} onClick={() => setKind(entry.id)}>{entry.label}（{data[entry.id]?.length || 0}）</Button>)}</div>
        <div className="flex flex-wrap items-center gap-2"><input ref={fileRef} type="file" accept=".xml,text/xml,application/xml" className="hidden" onChange={(event) => void importXML(event.target.files?.[0])} /><Button type="button" onClick={() => fileRef.current?.click()} isDisabled={busy}>上传 XML 导入</Button><Button type="button" variant="outline" onClick={add} isDisabled={busy}>新增条目</Button><Button type="button" className="bg-indigo-600 hover:bg-indigo-500" onClick={() => void save()} isDisabled={busy}>保存当前列表</Button></div>
        {status && <p className="text-sm text-indigo-300">{status}</p>}
        <div className="max-h-[520px] overflow-auto rounded-md border border-slate-800">
          <table className="w-full text-sm"><thead className="sticky top-0 bg-slate-800 text-slate-300"><tr><th className="p-2 text-left">ID</th><th className="p-2 text-left">名称</th><th className="p-2 text-right">操作</th></tr></thead><tbody>
            {items.map((item, index) => <tr key={`${kind}-${index}`} className="border-t border-slate-800"><td className="p-2"><input type="number" min="0" value={item.id} onChange={(event) => update(index, 'id', event.target.value)} className="w-28 rounded border border-slate-700 bg-slate-800 px-2 py-1 text-white" /></td><td className="p-2"><input value={item.name} onChange={(event) => update(index, 'name', event.target.value)} className="w-full rounded border border-slate-700 bg-slate-800 px-2 py-1 text-white" /></td><td className="p-2 text-right"><Button type="button" size="sm" variant="ghost" onClick={() => remove(index)}>删除</Button></td></tr>)}
            {!items.length && <tr><td colSpan={3} className="p-8 text-center text-slate-500">暂无数据，请上传 XML 或新增条目。</td></tr>}
          </tbody></table>
        </div>
      </CardContent>
    </Card>
  )
}
