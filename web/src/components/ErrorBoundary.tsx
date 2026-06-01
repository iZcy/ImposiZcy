import { Component, ReactNode, ErrorInfo } from 'react'
import { Button } from '@kzcy/dashboard'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export default class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('ErrorBoundary caught:', error, info.componentStack)
  }

  handleReload = () => {
    this.setState({ hasError: false, error: null })
    window.location.reload()
  }

  render() {
    if (this.state.hasError) {
      return (
        <div
          className="flex flex-col items-center justify-center min-h-[400px] gap-4"
          style={{ color: 'var(--text-secondary)' }}
        >
          <div className="text-4xl">Something went wrong</div>
          <p className="text-sm" style={{ color: 'var(--text-muted)', maxWidth: 480, textAlign: 'center' }}>
            {this.state.error?.message || 'An unexpected error occurred.'}
          </p>
          <Button onClick={this.handleReload}>Reload</Button>
        </div>
      )
    }
    return this.props.children
  }
}
