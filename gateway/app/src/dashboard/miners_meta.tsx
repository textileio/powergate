import React, {FunctionComponent} from 'react'
import Link from '@material-ui/core/Link'
import { makeStyles } from '@material-ui/core/styles'
import Table from '@material-ui/core/Table'
import TableBody from '@material-ui/core/TableBody'
import TableCell from '@material-ui/core/TableCell'
import TableHead from '@material-ui/core/TableHead'
import TableRow from '@material-ui/core/TableRow'
import Title from './title'

export type Item = {
  id: string
  online: boolean
  country: string
  latitude: number
  longitude: number
  userAgent: string
  lastUpdated: Date
}

type Props = {
  data: Item[]
}

const MinersMeta: FunctionComponent<Props> = (props) => {
  const rows = props.data.map(item => (
    <TableRow key={item.id}>
      <TableCell>{item.id}</TableCell>
      <TableCell>{item.online ? "true" : "false"}</TableCell>
      <TableCell>{item.country}</TableCell>
      <TableCell>{item.latitude}</TableCell>
      <TableCell>{item.longitude}</TableCell>
      <TableCell>{item.userAgent}</TableCell>
      <TableCell align="right">{item.lastUpdated.toLocaleString()}</TableCell>
    </TableRow>
  ))
  return (
    <React.Fragment>
      <Title>Miner Metadata</Title>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>ID</TableCell>
            <TableCell>Online</TableCell>
            <TableCell>Country</TableCell>
            <TableCell>Latitude</TableCell>
            <TableCell>Longitude</TableCell>
            <TableCell>User Agent</TableCell>
            <TableCell align="right">Last Updated</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {rows}
        </TableBody>
      </Table>
    </React.Fragment>
  )
}

export default MinersMeta
