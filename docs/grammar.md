$$
\begin{align}
  \text{Prog} &\to [\text{Stmt}]^* \\
  [\text{Stmt}] &\to
  \begin{cases}
    \text{import} \ \text{string\_lit} \\
    \text{fn} \ \text{ident} \ ([\text{ParamList}]?) \ [:\text{Type}]? \ [\text{Block}] \\
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
  [\text{ParamList}] &\to [\text{Param}] (,\ [\text{Param}])^* \\
  [\text{Param}] &\to \text{ident} [:\text{Type}]? [=\text{Expr}]? \\
  [\text{Type}] &\to [*]^* (\texttt{[}\ \text{int\_lit}\ \texttt{]})? \ (\text{i32} \mid \text{i64} \mid \text{f64} \mid \text{bool} \mid \text{string} \mid \text{int} \mid \text{float}) \\
  [\text{Expr}] &\to \text{fn} \ ([\text{ParamList}]?) \ [:\text{Type}]? \ [\text{Block}] \mid [\text{Cmp}] ([==|!=|<|>|<=|>=] [\text{Cmp}])? \\
  [\text{Cmp}] &\to [\text{Term}] ([+|-] [\text{Term}])^* \\
  [\text{Term}] &\to [\text{Factor}] ([*|/] [\text{Factor}])^* \\
  [\text{Factor}] &\to \text{int\_lit} \mid \text{ident} \mid \text{ident}([\text{Expr}] (,\ [\text{Expr}])^*)? \mid \text{ident}.\text{ident}([\text{Expr}] (,\ [\text{Expr}])^*)? \mid ([\text{Expr}])
\end{align}
$$
