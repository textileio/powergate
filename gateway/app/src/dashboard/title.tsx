import React, { FunctionComponent } from 'react'
import Typography from '@material-ui/core/Typography'

type Props = {}

const title: FunctionComponent<Props> = (props) => {
  return (
    <Typography component="h2" variant="h6" color="primary" gutterBottom>
      {props.children}
    </Typography>
  )
}

export default title