import React, {FunctionComponent} from 'react'
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
  title: React.ReactNode
  headers: React.ReactNode[]
  rows: React.ReactNode[][]
}

const WrappedTable: FunctionComponent<Props> = (props) => {
  const headerCells = props.headers.map((value, i) => <TableCell key={i}>{value}</TableCell>)
  const rows = props.rows.map((row, i) => {
    const cells = row.map((value, j) => <TableCell key={j}>{value}</TableCell>)
    return <TableRow key={i}>{cells}</TableRow>
  })
  return (
    <React.Fragment>
      <Title>{props.title}</Title>
      <Table size="small">
        <TableHead>
          <TableRow>
            {headerCells}
          </TableRow>
        </TableHead>
        <TableBody>
          {rows}
        </TableBody>
      </Table>
    </React.Fragment>
  )
}

export default WrappedTable
