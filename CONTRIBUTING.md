# Contributing to Lawrence OSS

Thank you for your interest in contributing to Lawrence OSS! We welcome contributions from the community and are grateful for your support.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Commit Messages](#commit-messages)
- [Getting Help](#getting-help)

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the problem
- **Expected behavior** vs **actual behavior**
- **Environment details** (OS, Docker version, Go version, etc.)
- **Logs and error messages** (sanitize sensitive data)
- **Screenshots** if applicable

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide detailed description** of the proposed feature
- **Explain why this enhancement would be useful**
- **List examples** of how it would work
- **Note any alternatives** you've considered

### Contributing Code

We welcome code contributions! Here are ways you can help:

- Fix bugs
- Implement new features
- Improve documentation
- Add tests
- Optimize performance
- Enhance UI/UX

### Improving Documentation

Documentation improvements are always welcome:

- Fix typos or clarify existing docs
- Add examples and tutorials
- Translate documentation
- Create diagrams or screenshots
- Write blog posts or guides

## Development Setup

### Prerequisites

- **Go 1.24+** - Backend development
- **Node.js 20+** - Frontend development
- **pnpm** - Frontend package manager
- **Docker** - For containerized builds
- **Make** - Build automation
- **Git** - Version control

### Getting Started

1. **Fork and clone the repository**

```bash
git clone https://github.com/YOUR_USERNAME/lawrence-oss.git
cd lawrence-oss
```

2. **Install Go dependencies**

```bash
go mod download
```

3. **Install frontend dependencies**

```bash
cd ui
pnpm install
cd ..
```

4. **Build the project**

```bash
make build
```

5. **Run locally**

```bash
# Terminal 1: Run backend
make run

# Terminal 2: Run frontend dev server
cd ui
pnpm dev
```

6. **Access the application**

- Backend API: http://localhost:8080
- Frontend Dev: http://localhost:5173
- Health Check: http://localhost:8080/health

### Development with Docker

```bash
# Build Docker image
make docker-build

# Run with Docker Compose
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down
```

## Pull Request Process

### Before Submitting

1. **Create an issue** first to discuss major changes
2. **Search existing PRs** to avoid duplicates
3. **Test your changes** thoroughly
4. **Update documentation** if needed
5. **Follow coding standards** (see below)

### Submission Steps

1. **Create a feature branch**

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

2. **Make your changes**

- Write clear, concise code
- Add tests for new functionality
- Update relevant documentation
- Follow existing code style

3. **Test your changes**

```bash
# Run backend tests
make test

# Run backend tests with coverage
make test-coverage

# Run frontend tests
cd ui
pnpm test

# Run linting
make lint
cd ui && pnpm lint

# Format code
make fmt
cd ui && pnpm format
```

4. **Commit your changes**

```bash
git add .
git commit -m "feat: add your feature description"
```

See [Commit Messages](#commit-messages) for guidelines.

5. **Push to your fork**

```bash
git push origin feature/your-feature-name
```

6. **Open a Pull Request**

- Use a clear, descriptive title
- Reference related issues (e.g., "Fixes #123")
- Describe what changed and why
- Include screenshots for UI changes
- Mark as draft if work-in-progress

### PR Review Process

- Maintainers will review your PR
- Address feedback and requested changes
- Keep your PR updated with main branch
- Once approved, maintainers will merge

### After Your PR is Merged

- Delete your feature branch
- Update your local main branch
- Celebrate! 🎉

## Coding Standards

### Go Backend

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for formatting (enforced by CI)
- Write meaningful comments for exported functions
- Keep functions small and focused
- Use structured logging (zap)
- Handle errors explicitly, don't ignore them

**Example:**

```go
// GetAgent retrieves an agent by ID from the store.
// Returns ErrAgentNotFound if the agent doesn't exist.
func (s *AgentService) GetAgent(ctx context.Context, id uuid.UUID) (*Agent, error) {
    agent, err := s.store.GetAgent(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get agent: %w", err)
    }
    return agent, nil
}
```

### TypeScript Frontend

- Use TypeScript for type safety
- Follow existing component patterns
- Use functional components with hooks
- Keep components small and reusable
- Write accessible UI (ARIA labels, keyboard navigation)
- Use Tailwind CSS for styling

**Example:**

```tsx
interface AgentCardProps {
  agent: Agent;
  onSelect?: (agent: Agent) => void;
}

export function AgentCard({ agent, onSelect }: AgentCardProps) {
  return (
    <div 
      className="rounded-lg border p-4 hover:bg-accent"
      onClick={() => onSelect?.(agent)}
      role="button"
      tabIndex={0}
    >
      <h3 className="font-semibold">{agent.name}</h3>
      <p className="text-sm text-muted-foreground">{agent.status}</p>
    </div>
  );
}
```

### General Guidelines

- **Be consistent** with existing code style
- **Write clear variable names** (no single letters except loops)
- **Add comments** for complex logic
- **Keep lines under 120 characters** when reasonable
- **Avoid premature optimization** - clarity first
- **Don't commit commented-out code** or debug logs

## Testing Guidelines

### Backend Testing

All new code should include tests:

```go
func TestGetAgent(t *testing.T) {
    // Setup
    store := &MockStore{}
    service := NewAgentService(store, logger)
    agentID := uuid.New()
    
    // Test
    agent, err := service.GetAgent(context.Background(), agentID)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, agent)
    assert.Equal(t, agentID, agent.ID)
}
```

### Frontend Testing

Write tests for components and hooks:

```tsx
import { render, screen } from '@testing-library/react';
import { AgentCard } from './AgentCard';

describe('AgentCard', () => {
  it('renders agent name', () => {
    const agent = { id: '1', name: 'Test Agent', status: 'online' };
    render(<AgentCard agent={agent} />);
    expect(screen.getByText('Test Agent')).toBeInTheDocument();
  });
});
```

### Test Requirements

- **Unit tests** for business logic
- **Integration tests** for API endpoints
- **E2E tests** for critical user flows
- **Aim for >70% coverage** for new code
- **Test both success and error cases**

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/) for clear history:

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation only
- `style:` - Code style (formatting, no logic change)
- `refactor:` - Code refactoring
- `perf:` - Performance improvement
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks
- `ci:` - CI/CD changes

### Examples

```
feat(opamp): add support for custom agent attributes

Adds ability to send custom attributes with OpAMP connections.
This allows agents to send environment-specific metadata.

Closes #123
```

```
fix(api): handle nil pointer in agent handler

Fixes panic when agent disconnects during config update.
Added nil checks and proper error handling.

Fixes #456
```

```
docs: update deployment guide with TLS setup

Added section on configuring TLS with reverse proxy.
Includes nginx and Caddy examples.
```

### Commit Guidelines

- **Use present tense** ("add feature" not "added feature")
- **Keep subject under 72 characters**
- **Capitalize subject line**
- **No period at end of subject**
- **Reference issues** in footer
- **Explain "why"** in body, not just "what"

## Getting Help

### Questions?

- **GitHub Issues** - For bugs and feature requests
- **GitHub Discussions** - For questions and general discussion
- **Documentation** - Check the [docs/](docs/) directory

### Development Questions

If you're stuck or need clarification:

1. Check existing issues and discussions
2. Review the documentation
3. Ask in GitHub Discussions
4. Be specific about your problem
5. Include relevant code snippets

### Need a Maintainer?

Tag `@getlawrence` in issues or PRs if:

- You need a review
- You're blocked on a decision
- Your PR has been waiting >5 days
- You found a security issue (see [SECURITY.md](SECURITY.md))

## Recognition

Contributors will be:

- Listed in release notes
- Added to GitHub contributors
- Mentioned in project documentation
- Our eternal gratitude! 🙏

## License

By contributing to Lawrence OSS, you agree that your contributions will be licensed under the [Apache 2.0 License](LICENSE).

---

**Thank you for contributing to Lawrence OSS!** Every contribution, no matter how small, makes a difference. We're excited to work with you! 🚀

