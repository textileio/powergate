import React from 'react'
import CssBaseline from '@material-ui/core/CssBaseline'

import {APIClient} from './_proto/ask_pb_service'
import {GetRequest} from './_proto/ask_pb'

import Dashboard from './dashboard/Dashboard'

export default function App() {
  return (
    <React.Fragment>
      <CssBaseline />
      {/* The rest of your application */}
      <Dashboard />
    </React.Fragment>
  )
}
