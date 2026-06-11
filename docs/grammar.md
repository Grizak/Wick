$$
\begin{align}
  \text{Prog} &\to [\text{Stmt}]^* \\
  [\text{Stmt}] &\to
  \begin{cases}
    [\text{Expr}] \\
    [\text{Cmp}] \\
    [\text{Term}] \\
    [\text{Factor}] \\
  \ldots \\
  \text{for} \ [\text{Expr}] \ [\text{Block}] \\
  \text{for} \ [\text{Stmt}] ; [\text{Expr}] ; [\text{Stmt}] \ [\text{Block}] \\
  \text{for} \ [\text{Block}] \\
  \text{if} \ [\text{Expr}] \ [\text{Block}] \ [\text{else} \ [\text{Block}]]?
  \end{cases} \\
  [\text{Block}] &\to \{ [\text{Stmt}]^* \} \\
  [\text{Expr}] &\to [\text{Cmp}] ([==|!=|<|>|<=|>=] [\text{Cmp}])? \\
  [\text{Cmp}] &\to [\text{Term}] ([+|-] [\text{Term}])^* \\
  [\text{Term}] &\to [\text{Factor}] ([*|/] [\text{Factor}])^* \\
  [\text{Factor}] &\to \text{int\_lit} \mid \text{ident} \mid ([\text{Expr}])
\end{align}
$$
