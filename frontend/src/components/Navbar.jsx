import { NavLink } from "react-router-dom";
import "./Navbar.css";

export default function Navbar({ onLogout }) {
  return (
    <>
      <header className="navbar-header">
        <h1>Dentist Clinic</h1>
      </header>

      <nav className="navbar-menu" role="navigation" aria-label="Main navigation">
        <NavLink to="/create" className={({ isActive }) => (isActive ? "active" : "")}>Create</NavLink>
        <NavLink to="/live" className={({ isActive }) => (isActive ? "active" : "")}>Live Monitor</NavLink>
        <NavLink to="/history" className={({ isActive }) => (isActive ? "active" : "")}>History</NavLink>
        <button className="navbar-logout" type="button" onClick={onLogout}>Log out</button>
      </nav>
    </>
  );
}