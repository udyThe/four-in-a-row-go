# Contributing Guide

Thank you for your interest in contributing to the 4 in a Row project! This document provides guidelines for contributing.

## 📋 Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)

## 📜 Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person
- Help others learn and grow

## 🚀 Getting Started

1. **Fork the repository**
2. **Clone your fork**
   ```bash
   git clone https://github.com/your-username/4-in-a-row.git
   cd 4-in-a-row
   ```
3. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/original-owner/4-in-a-row.git
   ```

## 💻 Development Setup

### Backend Development

```bash
cd backend

# Install dependencies
go mod download

# Run tests
go test ./...

# Run locally
go run main.go
```

### Frontend Development

```bash
cd frontend

# Install dependencies
npm install

# Run development server
npm start

# Run tests
npm test

# Build for production
npm run build
```

### Full Stack with Docker

```bash
# Start all services
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## 🔨 Making Changes

### 1. Create a Branch

```bash
# Update your fork
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes

Follow these guidelines:
- Write clear, concise code
- Add comments for complex logic
- Update documentation if needed
- Add tests for new features
- Ensure existing tests pass

### 3. Commit Your Changes

```bash
# Stage changes
git add .

# Commit with meaningful message
git commit -m "Add: Brief description of changes"
```

**Commit Message Format:**
- `Add:` New feature
- `Fix:` Bug fix
- `Update:` Changes to existing feature
- `Refactor:` Code refactoring
- `Docs:` Documentation changes
- `Test:` Test additions/changes

## 🧪 Testing

### Backend Tests

```bash
cd backend
go test ./... -v
go test ./... -cover
```

### Frontend Tests

```bash
cd frontend
npm test
npm run test:coverage
```

### Integration Tests

```bash
# Start services
docker-compose up -d

# Run integration tests
# (Add your integration test commands here)

# Stop services
docker-compose down
```

### Manual Testing

Refer to [TESTING_CHECKLIST.md](TESTING_CHECKLIST.md) for comprehensive testing.

## 📤 Submitting Changes

### 1. Push to Your Fork

```bash
git push origin feature/your-feature-name
```

### 2. Create Pull Request

1. Go to GitHub repository
2. Click "New Pull Request"
3. Select your branch
4. Fill in PR template:
   - **Title**: Clear, concise description
   - **Description**: What changed and why
   - **Related Issues**: Link any related issues
   - **Testing**: Describe how you tested
   - **Screenshots**: If UI changes

### 3. PR Review Process

- Automated checks must pass (CI/CD)
- Code review by maintainers
- Address feedback and requested changes
- Squash commits if requested
- Merge when approved

## 🎨 Code Style

### Go Code Style

Follow standard Go conventions:
- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use meaningful variable names
- Keep functions focused and small
- Add godoc comments for exported functions

Example:
```go
// CalculateScore computes the score for a given board position
func CalculateScore(board *Board, player Player) int {
    // Implementation
}
```

### JavaScript/React Code Style

- Use ES6+ syntax
- Use functional components with hooks
- Use meaningful component and variable names
- Keep components focused and reusable
- Add JSDoc comments for complex functions

Example:
```javascript
/**
 * Calculates the winner of the game
 * @param {Array} board - The game board
 * @returns {number|null} The winning player or null
 */
const calculateWinner = (board) => {
    // Implementation
};
```

### CSS Style

- Use BEM naming convention
- Keep selectors specific but not overly nested
- Use CSS variables for colors and common values
- Mobile-first responsive design

## 📝 Documentation

### Code Documentation

- Add comments for complex logic
- Document public APIs
- Update README for new features
- Add JSDoc/Godoc for functions

### User Documentation

Update relevant docs:
- README.md - Main documentation
- ARCHITECTURE.md - System design changes
- DEPLOYMENT.md - Deployment changes
- QUICK_START.md - Quick reference updates

## 🐛 Bug Reports

### Before Submitting

- Check if bug already reported
- Verify bug exists in latest version
- Collect relevant information

### Bug Report Template

```markdown
## Bug Description
Clear description of the bug

## Steps to Reproduce
1. Step 1
2. Step 2
3. Step 3

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: [e.g., Windows 10]
- Browser: [e.g., Chrome 96]
- Version: [e.g., 1.0.0]

## Screenshots
If applicable

## Additional Context
Any other relevant information
```

## 💡 Feature Requests

### Feature Request Template

```markdown
## Feature Description
Clear description of the feature

## Use Case
Why is this feature needed?

## Proposed Solution
How should it work?

## Alternatives Considered
Other approaches you considered

## Additional Context
Any other relevant information
```

## 🔍 Code Review Guidelines

### As a Reviewer

- Be respectful and constructive
- Focus on code quality and correctness
- Suggest improvements, don't demand
- Approve when ready
- Test the changes if possible

### As a Contributor

- Respond to feedback promptly
- Ask questions if unclear
- Make requested changes
- Update PR description if scope changes
- Be patient and respectful

## 📦 Release Process

### Version Numbering

Follow Semantic Versioning (SemVer):
- **Major** (1.0.0): Breaking changes
- **Minor** (0.1.0): New features, backward compatible
- **Patch** (0.0.1): Bug fixes

### Release Checklist

1. Update version numbers
2. Update CHANGELOG.md
3. Run full test suite
4. Build production images
5. Tag release
6. Deploy to production
7. Update documentation

## 🏷️ Issue Labels

- `bug` - Something isn't working
- `enhancement` - New feature or request
- `documentation` - Documentation improvements
- `good first issue` - Good for newcomers
- `help wanted` - Extra attention needed
- `question` - Further information requested
- `wontfix` - This will not be worked on

## 🤝 Community

### Getting Help

- Check documentation first
- Search existing issues
- Ask in discussions
- Create new issue if needed

### Staying Updated

- Watch repository for notifications
- Follow release notes
- Join discussions
- Contribute to conversations

## 📊 Contribution Stats

We appreciate all contributions:
- Code contributions
- Documentation improvements
- Bug reports
- Feature suggestions
- Code reviews
- Community support

## 🙏 Recognition

All contributors will be:
- Listed in CONTRIBUTORS.md
- Mentioned in release notes
- Appreciated in the community

Thank you for contributing! 🎉
