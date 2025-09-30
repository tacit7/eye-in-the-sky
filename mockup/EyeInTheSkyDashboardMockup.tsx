import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Activity, Users, Clock } from "lucide-react"

export default function EyeInTheSkyDashboardMockup() {
  return (
    <div className="min-h-screen bg-[#1a1b1e] text-gray-100 p-8 space-y-8">
      <header className="flex items-center justify-between bg-primary-800 p-4 rounded-xl shadow-md">
        <h1 className="text-xl font-semibold text-white tracking-wide">Eye in the Sky</h1>
        <Button className="bg-primary-500 hover:bg-primary-600 text-white">Launch Agent</Button>
      </header>

      <section className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Card className="bg-[#232427] border border-gray-700">
          <CardContent className="p-4 flex flex-col items-center">
            <Users className="text-primary-400 w-6 h-6 mb-2" />
            <p className="text-sm text-gray-400">Active Agents</p>
            <p className="text-2xl font-semibold text-primary-400">3</p>
          </CardContent>
        </Card>
        <Card className="bg-[#232427] border border-gray-700">
          <CardContent className="p-4 flex flex-col items-center">
            <Activity className="text-primary-400 w-6 h-6 mb-2" />
            <p className="text-sm text-gray-400">Sessions Today</p>
            <p className="text-2xl font-semibold text-primary-400">3</p>
          </CardContent>
        </Card>
        <Card className="bg-[#232427] border border-gray-700">
          <CardContent className="p-4 flex flex-col items-center">
            <Clock className="text-primary-400 w-6 h-6 mb-2" />
            <p className="text-sm text-gray-400">Inactive (&gt;2h)</p>
            <p className="text-2xl font-semibold text-primary-400">0</p>
          </CardContent>
        </Card>
      </section>

      <section className="bg-[#232427] border border-gray-700 rounded-xl shadow-lg overflow-hidden">
        <div className="p-4 border-b border-gray-700 flex justify-between items-center">
          <h2 className="text-lg font-semibold">Active Claude Code Agents</h2>
          <span className="text-sm text-primary-400 font-medium">3 Active</span>
        </div>

        <div className="divide-y divide-gray-700">
          {[1, 2, 3].map((id) => (
            <div key={id} className="flex justify-between items-center p-4 hover:bg-gray-800 transition-colors">
              <div>
                <p className="font-medium text-gray-100">Agent {id}</p>
                <p className="text-sm text-gray-400">Working on UI improvements for Eye in the Sky</p>
              </div>
              <div className="flex gap-2">
                <Button variant="outline" className="border border-primary-500 text-primary-400 hover:bg-primary-600 hover:text-white">View</Button>
                <Button className="bg-primary-600 hover:bg-primary-700 text-white">End</Button>
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}