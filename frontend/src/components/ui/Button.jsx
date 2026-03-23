export function Button({
  children,
  variant  = 'primary',
  size     = 'md',
  loading  = false,
  disabled = false,
  icon,
  type     = 'button',
  onClick,
  className = '',
  ...rest
}) {
  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled || loading}
      className={`btn btn-${variant} btn-${size} ${className}`}
      {...rest}
    >
      {loading ? <span className="btn-spinner" /> : icon}
      {children}
    </button>
  )
}

