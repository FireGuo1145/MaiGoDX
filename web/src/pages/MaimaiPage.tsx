import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { RANK_SUMMARY, type Stats } from '@/types'

interface MaimaiPageProps {
  stats: Stats | null
}

export function MaimaiPage({ stats }: MaimaiPageProps) {
  const plays = stats?.recentPlays || []

  return (
    <div className="space-y-6">
      <Tabs defaultSelectedKey="recent" className="w-full">
        <TabsList className="bg-slate-900 border border-slate-800">
          <TabsTrigger id="recent" className="data-[selected]:bg-indigo-600">Recent Plays</TabsTrigger>
          <TabsTrigger id="stats" className="data-[selected]:bg-indigo-600">Statistics</TabsTrigger>
        </TabsList>

        <TabsContent id="recent" className="mt-6">
          <Card className="bg-slate-900 border-slate-800 overflow-hidden">
            <Table>
              <TableHeader className="bg-slate-800/50">
                <TableRow>
                  <TableHead className="text-slate-300">Music</TableHead>
                  <TableHead className="text-slate-300">Level</TableHead>
                  <TableHead className="text-slate-300 text-right">Achievement</TableHead>
                  <TableHead className="text-slate-300 text-right">Score</TableHead>
                  <TableHead className="text-slate-300 text-right">Date</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {plays.length ? plays.map((play) => (
                  <TableRow key={play.ID} className="border-slate-800 hover:bg-slate-800/30">
                    <TableCell className="font-bold text-white">Song ID: {play.musicId}</TableCell>
                    <TableCell><Badge variant="outline" className="border-indigo-500/50 text-indigo-400">LV.{play.level}</Badge></TableCell>
                    <TableCell className="text-right font-mono text-emerald-400">{(play.achievement / 10000).toFixed(4)}%</TableCell>
                    <TableCell className="text-right font-mono">{play.score.toLocaleString()}</TableCell>
                    <TableCell className="text-right text-slate-500 text-xs">{play.createDate || '—'}</TableCell>
                  </TableRow>
                )) : (
                  <TableRow><TableCell colSpan={5} className="text-center py-12 text-slate-500">No play history found.</TableCell></TableRow>
                )}
              </TableBody>
            </Table>
          </Card>
        </TabsContent>

        <TabsContent id="stats" className="mt-6">
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4">
            {RANK_SUMMARY.map((rank) => (
              <Card key={rank} className="bg-slate-900 border-slate-800 text-center p-6">
                <p className="text-2xl font-black text-indigo-400 mb-1">{rank}</p>
                <p className="text-sm text-slate-500 font-bold">124 songs</p>
              </Card>
            ))}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
