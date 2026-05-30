$$
\begin{align}
  \text{Prog} &\to [\text{Stmt}]^* \\
  [\text{Stmt}] &\to
  \begin{cases}
    \text{exit}([\text{Expr}])
  \end{cases} \\
  [\text{Expr}] &\to [\text{Term}] ([+|-] [\text{Term}])^* \\
  [\text{Term}] &\to [\text{Factor}] ([*|/] [\text{Factor}])^* \\
  [\text{Factor}] &\to \text{int\_lit} \mid ([\text{Expr}])
\end{align}
$$
