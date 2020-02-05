import React from 'react'
import Button from '@material-ui/core/Button'
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

function getAsks() {
  const r = new GetRequest()
  const c = new APIClient("http://40.117.82.59:6002")
  const resp = c.get(r, (err, resp) => {
    if (err) {
      console.log("got error: ", err)
    }
    if (resp) {
      console.log("got response: ", resp.getIndex()?.toObject())
      const foo = resp.toObject()
    }
  })
}

getAsks()
