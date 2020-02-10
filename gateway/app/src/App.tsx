import React, {useEffect} from 'react'
import CssBaseline from '@material-ui/core/CssBaseline'
import Client from './client'
import Dashboard from './dashboard/Dashboard'

export default function App() {

  useEffect(() => {
    Client.initialize("http://40.117.82.59:6002")
  }, [])

  return (
    <React.Fragment>
      <CssBaseline />
      {/* The rest of your application */}
      <Dashboard />
    </React.Fragment>
  )
}
