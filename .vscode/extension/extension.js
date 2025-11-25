const vscode = require('vscode');
const { exec } = require('child_process');
const path = require('path');

function activate(context) {
    console.log('Go Ollama Helper extension activated');

    // Explain current file
    let explainFile = vscode.commands.registerCommand('go-ollama-helper.explainFile', async () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) {
            vscode.window.showErrorMessage('No active editor');
            return;
        }

        const filePath = editor.document.fileName;
        if (!filePath.endsWith('.go')) {
            vscode.window.showErrorMessage('Not a Go file');
            return;
        }

        vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "🤖 Analyzing Go code with Ollama...",
            cancellable: false
        }, (progress) => {
            return new Promise((resolve) => {
                exec(`go-ollama-agent explain "${filePath}"`, (error, stdout, stderr) => {
                    if (error) {
                        vscode.window.showErrorMessage(`Ollama Error: ${error.message}`);
                        resolve();
                        return;
                    }

                    // Show result in new editor
                    vscode.workspace.openTextDocument({
                        content: `# Ollama Analysis: ${path.basename(filePath)}\n\n${stdout}`,
                        language: 'markdown'
                    }).then(doc => {
                        vscode.window.showTextDocument(doc, vscode.ViewColumn.Beside);
                    });

                    resolve();
                });
            });
        });
    });

    // Generate handler
    let generateHandler = vscode.commands.registerCommand('go-ollama-helper.generateHandler', async () => {
        const handlerName = await vscode.window.showInputBox({
            prompt: 'Enter handler name',
            placeHolder: 'e.g., UserCreate, ProductList'
        });

        if (!handlerName) return;

        const purpose = await vscode.window.showInputBox({
            prompt: 'Enter handler purpose',
            placeHolder: 'e.g., Handle user registration, List products'
        });

        if (!purpose) return;

        vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: `🤖 Generating ${handlerName} handler...`,
            cancellable: false
        }, (progress) => {
            return new Promise((resolve) => {
                exec(`go-ollama-agent generate handler "${handlerName}" "${purpose}"`, (error, stdout, stderr) => {
                    if (error) {
                        vscode.window.showErrorMessage(`Ollama Error: ${error.message}`);
                        resolve();
                        return;
                    }

                    // Create and show generated code
                    vscode.workspace.openTextDocument({
                        content: `// Generated handler: ${handlerName}\n// Purpose: ${purpose}\n\n${stdout}`,
                        language: 'go'
                    }).then(doc => {
                        vscode.window.showTextDocument(doc, vscode.ViewColumn.Beside);
                    });

                    resolve();
                });
            });
        });
    });

    // Generate tests
    let generateTests = vscode.commands.registerCommand('go-ollama-helper.generateTests', async () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) return;

        const filePath = editor.document.fileName;
        
        vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "🧪 Generating tests with Ollama...",
            cancellable: false
        }, (progress) => {
            return new Promise((resolve) => {
                exec(`go-ollama-agent test "${filePath}"`, (error, stdout, stderr) => {
                    if (error) {
                        vscode.window.showErrorMessage(`Ollama Error: ${error.message}`);
                        resolve();
                        return;
                    }

                    // Create test file
                    const testFilePath = filePath.replace('.go', '_test.go');
                    vscode.workspace.openTextDocument({
                        content: stdout,
                        language: 'go'
                    }).then(doc => {
                        vscode.window.showTextDocument(doc, vscode.ViewColumn.Beside);
                    });

                    resolve();
                });
            });
        });
    });

    // Debug code
    let debugCode = vscode.commands.registerCommand('go-ollama-helper.debugCode', async () => {
        const errorMsg = await vscode.window.showInputBox({
            prompt: 'Enter error message to debug',
            placeHolder: 'e.g., nil pointer, index out of range'
        });

        if (!errorMsg) return;

        const editor = vscode.window.activeTextEditor;
        if (!editor) return;

        const filePath = editor.document.fileName;
        
        vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: "🐛 Debugging with Ollama...",
            cancellable: false
        }, (progress) => {
            return new Promise((resolve) => {
                exec(`go-ollama-agent debug "${errorMsg}" "${filePath}"`, (error, stdout, stderr) => {
                    if (error) {
                        vscode.window.showErrorMessage(`Ollama Error: ${error.message}`);
                        resolve();
                        return;
                    }

                    vscode.workspace.openTextDocument({
                        content: `# Debug Analysis\n\n${stdout}`,
                        language: 'markdown'
                    }).then(doc => {
                        vscode.window.showTextDocument(doc, vscode.ViewColumn.Beside);
                    });

                    resolve();
                });
            });
        });
    });

    context.subscriptions.push(explainFile, generateHandler, generateTests, debugCode);
}

function deactivate() {}

module.exports = { activate, deactivate };