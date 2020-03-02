import React, {useEffect, FunctionComponent, useState} from 'react'
import CssBaseline from '@material-ui/core/CssBaseline'
import Client from './client'
import Dashboard from './dashboard/Dashboard'
import {Index as MinerIndex} from './_proto/miner_pb'
import {Index as AskIndex} from './_proto/ask_pb'
import {Index as SlashingIndex} from './_proto/slashing_pb'
import {MinerScore} from './_proto/reputation_pb'

const App: FunctionComponent = () => {

  const [minersIndex, setMinerIndex] = useState<MinerIndex.AsObject|undefined>(undefined)
  const [askIndex, setAskIndex] = useState<AskIndex.AsObject|undefined>(undefined)
  const [slashingIndex, setSlashingIndex] = useState<SlashingIndex.AsObject|undefined>(undefined)
  const [topMiners, setTopMiners] = useState<MinerScore.AsObject[]|undefined>(undefined)

  useEffect(() => {
    Client.initialize("http://40.117.82.59:6002")
  }, [])

  useEffect(() => {
    const fetchData = async () => {
      const i = await Client.shared().miners.get()
      setMinerIndex(i)
    }
    fetchData()
  }, [])

  useEffect(() => {
    const fetchData = async () => {
      const i = await Client.shared().asks.get()
      setAskIndex(i)
    }
    fetchData()
  }, [])

  useEffect(() => {
    const fetchData = async () => {
      const i = await Client.shared().slashing.get()
      setSlashingIndex(i)
    }
    fetchData()
  }, [])

  useEffect(() => {
    const fetchData = async () => {
      const i = await Client.shared().reputation.getTopMiners()
      setTopMiners(i)
    }
    fetchData()
  }, [])

  return (
    <React.Fragment>
      <CssBaseline />
      {/* The rest of your application */}
      <Dashboard askIndex={askIndex} minerIndex={minersIndex} slashingIndex={slashingIndex} topMiners={topMiners} />
    </React.Fragment>
  )
}

export default App
